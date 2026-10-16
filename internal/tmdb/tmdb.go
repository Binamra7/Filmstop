// Package tmdb searches The Movie Database, with an in-memory cache so repeat
// queries (and every keystroke that lands on the same prefix) never leave the box.
package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Binamra7/filmstop/internal/store"
)

var ErrNoKey = errors.New("TMDB_API_KEY is not set")

type Client struct {
	key  string
	http *http.Client

	mu    sync.Mutex
	cache map[string]entry
}

type entry struct {
	movies []store.Movie
	at     time.Time
}

const (
	cacheTTL  = 30 * time.Minute
	cacheSize = 2000
	maxHits   = 8
)

// New accepts either a v3 API key or a v4 read access token.
func New(key string) *Client {
	return &Client{
		key:   strings.TrimSpace(key),
		http:  &http.Client{Timeout: 4 * time.Second},
		cache: make(map[string]entry),
	}
}

func (c *Client) Enabled() bool { return c.key != "" }

// Search finds films by title. fresh reports that the results came from TMDB
// rather than the cache, so the caller knows they haven't been stored yet.
func (c *Client) Search(ctx context.Context, query string) (movies []store.Movie, fresh bool, err error) {
	if c.key == "" {
		return nil, false, ErrNoKey
	}
	q := strings.ToLower(strings.Join(strings.Fields(query), " "))
	if q == "" {
		return nil, false, nil
	}
	c.mu.Lock()
	if e, ok := c.cache[q]; ok && time.Since(e.at) < cacheTTL {
		c.mu.Unlock()
		return e.movies, false, nil
	}
	c.mu.Unlock()

	var body struct {
		Results []struct {
			ID          int64  `json:"id"`
			Title       string `json:"title"`
			ReleaseDate string `json:"release_date"`
			PosterPath  string `json:"poster_path"`
			Backdrop    string `json:"backdrop_path"`
			Overview    string `json:"overview"`
		} `json:"results"`
	}
	v := url.Values{"query": {q}, "include_adult": {"false"}, "language": {"en-US"}, "page": {"1"}}
	if err := c.get(ctx, "/search/movie", v, &body); err != nil {
		return nil, false, err
	}
	movies = make([]store.Movie, 0, maxHits)
	for _, r := range body.Results {
		if len(movies) == maxHits {
			break
		}
		year := ""
		if len(r.ReleaseDate) >= 4 {
			year = r.ReleaseDate[:4]
		}
		movies = append(movies, store.Movie{ID: r.ID, Title: r.Title, Year: year, PosterPath: r.PosterPath, BackdropPath: r.Backdrop, Overview: r.Overview})
	}

	c.mu.Lock()
	if len(c.cache) >= cacheSize {
		// Cheap eviction: drop everything stale, and if that frees nothing, start over.
		for k, e := range c.cache {
			if time.Since(e.at) >= cacheTTL {
				delete(c.cache, k)
			}
		}
		if len(c.cache) >= cacheSize {
			clear(c.cache)
		}
	}
	c.cache[q] = entry{movies: movies, at: time.Now()}
	c.mu.Unlock()
	return movies, true, nil
}

func (c *Client) get(ctx context.Context, path string, v url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.themoviedb.org/3"+path, nil)
	if err != nil {
		return err
	}
	// v4 tokens are JWTs; v3 keys are 32 hex characters.
	if strings.Count(c.key, ".") == 2 {
		req.Header.Set("Authorization", "Bearer "+c.key)
	} else {
		v.Set("api_key", c.key)
	}
	req.URL.RawQuery = v.Encode()
	req.Header.Set("Accept", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		// A *url.Error quotes the full URL, which can carry the api_key.
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return fmt.Errorf("tmdb %s: %w", path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("tmdb %s: %s", path, res.Status)
	}
	return json.NewDecoder(res.Body).Decode(out)
}

// Details fetches the tagline, director, runtime and genres for one film, plus
// its backdrop and poster (search results sometimes lack ones the film page has).
func (c *Client) Details(ctx context.Context, id int64) (backdrop, poster string, d store.Details, err error) {
	if c.key == "" {
		return "", "", d, ErrNoKey
	}
	var body struct {
		Tagline  string `json:"tagline"`
		Runtime  int    `json:"runtime"`
		Backdrop string `json:"backdrop_path"`
		Poster   string `json:"poster_path"`
		Genres   []struct {
			Name string `json:"name"`
		} `json:"genres"`
		Credits struct {
			Crew []struct {
				Name string `json:"name"`
				Job  string `json:"job"`
			} `json:"crew"`
		} `json:"credits"`
	}
	v := url.Values{"append_to_response": {"credits"}, "language": {"en-US"}}
	if err := c.get(ctx, fmt.Sprintf("/movie/%d", id), v, &body); err != nil {
		return "", "", d, err
	}
	d = store.Details{Tagline: body.Tagline, Runtime: body.Runtime, Detailed: true}
	var directors, genres []string
	for _, p := range body.Credits.Crew {
		if p.Job == "Director" && len(directors) < 2 {
			directors = append(directors, p.Name)
		}
	}
	for _, g := range body.Genres {
		if len(genres) < 2 {
			genres = append(genres, g.Name)
		}
	}
	d.Director = strings.Join(directors, " & ")
	d.Genres = strings.Join(genres, ", ")
	return body.Backdrop, body.Poster, d, nil
}
