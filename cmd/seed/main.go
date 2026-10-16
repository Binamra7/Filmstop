// Command seed fills a local database with simulated ballots over real TMDB films,
// so the list has something on it during development. Never run it against
// production: the ballots are made up.
//
//	TMDB_API_KEY=... go run ./cmd/seed -db filmstop.db -voters 400
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/Binamra7/filmstop/internal/rank"
	"github.com/Binamra7/filmstop/internal/store"
)

type film struct {
	store.Movie
	appeal float64
}

func main() {
	dbPath := flag.String("db", "filmstop.db", "SQLite file (ignored when TURSO_DATABASE_URL is set)")
	voters := flag.Int("voters", 400, "simulated voters")
	flag.Parse()
	key := strings.TrimSpace(os.Getenv("TMDB_API_KEY"))
	if key == "" {
		log.Fatal("TMDB_API_KEY is required")
	}
	ctx := context.Background()

	films, err := topRated(key, 5)
	if err != nil {
		log.Fatal(err)
	}
	if len(films) < rank.Slots {
		log.Fatalf("TMDB returned %d films; a ballot needs %d", len(films), rank.Slots)
	}
	st, err := store.Open(store.Source(*dbPath))
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()
	movies := make([]store.Movie, len(films))
	for i, f := range films {
		movies[i] = f.Movie
	}
	if err := st.UpsertMovies(ctx, movies); err != nil {
		log.Fatal(err)
	}

	// Each voter draws five films without replacement, with probability
	// proportional to appeal × personal taste noise (a Plackett–Luce sampler via
	// the Gumbel trick).
	r := rand.New(rand.NewPCG(7, 11))
	for v := range *voters {
		type keyed struct {
			id int64
			k  float64
		}
		ks := make([]keyed, len(films))
		for i, f := range films {
			ks[i] = keyed{f.ID, math.Log(f.appeal) - math.Log(-math.Log(r.Float64()))}
		}
		sort.Slice(ks, func(i, j int) bool { return ks[i].k > ks[j].k })
		var b store.Ballot
		for i := range b {
			b[i] = ks[i].id
		}
		if err := st.Cast(ctx, fmt.Sprintf("seed%028d", v), b); err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("seeded %d ballots over %d films", *voters, len(films))
}

func topRated(key string, pages int) ([]film, error) {
	var out []film
	for p := 1; p <= pages; p++ {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.themoviedb.org/3/movie/top_rated?language=en-US&page=%d", p), nil)
		if err != nil {
			return nil, err
		}
		// Accept a v4 read token as well as a v3 key, as the server does.
		if strings.Count(key, ".") == 2 {
			req.Header.Set("Authorization", "Bearer "+key)
		} else {
			req.URL.RawQuery += "&api_key=" + url.QueryEscape(key)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			// A *url.Error quotes the full URL, which can carry the api_key.
			var ue *url.Error
			if errors.As(err, &ue) {
				err = ue.Err
			}
			return nil, fmt.Errorf("tmdb top_rated: %w", err)
		}
		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			return nil, fmt.Errorf("tmdb top_rated: %s", res.Status)
		}
		var body struct {
			Results []struct {
				ID          int64   `json:"id"`
				Title       string  `json:"title"`
				ReleaseDate string  `json:"release_date"`
				PosterPath  string  `json:"poster_path"`
				Backdrop    string  `json:"backdrop_path"`
				Overview    string  `json:"overview"`
				VoteAverage float64 `json:"vote_average"`
				VoteCount   float64 `json:"vote_count"`
			} `json:"results"`
		}
		err = json.NewDecoder(res.Body).Decode(&body)
		res.Body.Close()
		if err != nil {
			return nil, err
		}
		for _, m := range body.Results {
			year := ""
			if len(m.ReleaseDate) >= 4 {
				year = m.ReleaseDate[:4]
			}
			appeal := math.Exp(2*(m.VoteAverage-8)) * math.Log10(m.VoteCount+10)
			out = append(out, film{store.Movie{ID: m.ID, Title: m.Title, Year: year, PosterPath: m.PosterPath, BackdropPath: m.Backdrop, Overview: m.Overview}, appeal})
		}
	}
	return out, nil
}
