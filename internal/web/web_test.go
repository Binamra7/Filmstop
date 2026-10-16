package web

import (
	"context"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Binamra7/filmstop/internal/store"
	"github.com/Binamra7/filmstop/internal/tmdb"
	"github.com/gin-gonic/gin"
)

func TestBallotFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	titles := []string{"Stalker", "Yi Yi", "Heat", "Paris, Texas", "In the Mood for Love", "Ran"}
	var ms []store.Movie
	for i, title := range titles {
		ms = append(ms, store.Movie{ID: int64(i + 1), Title: title})
	}
	st.UpsertMovies(ctx, ms)
	srv, err := New(ctx, st, tmdb.New(""))
	if err != nil {
		t.Fatal(err)
	}
	h := srv.Handler()

	var cookie *http.Cookie
	do := func(method, path string, form url.Values, htmx bool) *httptest.ResponseRecorder {
		var req *http.Request
		if form != nil {
			req = httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else {
			req = httptest.NewRequest(method, path, nil)
		}
		if htmx {
			req.Header.Set("HX-Request", "true")
		}
		if cookie != nil {
			req.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		for _, c := range w.Result().Cookies() {
			if c.Name == voterCookie {
				cookie = c
			}
		}
		return w
	}

	if w := do("GET", "/", nil, false); w.Code != 200 || !strings.Contains(w.Body.String(), "No ballots yet") {
		t.Fatalf("home: %d", w.Code)
	}
	for id := 1; id <= 5; id++ {
		w := do("POST", "/ballot/pick", url.Values{"id": {digit(id)}}, true)
		if !strings.Contains(w.Body.String(), html.EscapeString(titles[id-1])) || !strings.Contains(w.Body.String(), "A discovery.") {
			t.Fatalf("pick %d: %s", id, w.Body.String())
		}
	}
	if w := do("POST", "/ballot/pick", url.Values{"id": {"6"}}, true); !strings.Contains(w.Body.String(), "full") {
		t.Fatal("sixth pick should be refused")
	}
	if w := do("POST", "/ballot/pick", url.Values{"id": {"999"}}, true); !strings.Contains(w.Body.String(), "in our list") {
		t.Fatal("unknown film should be refused")
	}
	do("POST", "/ballot/move", url.Values{"pos": {"5"}, "dir": {"up"}}, true) // swap 4 and 5
	if w := do("POST", "/ballot/cast", url.Values{}, false); w.Code != http.StatusSeeOther {
		t.Fatalf("cast: %d", w.Code)
	}
	b, _ := st.Draft(ctx, func() string { id, _ := srv.voters.verify(cookie.Value); return id }())
	if b != (store.Ballot{1, 2, 3, 5, 4}) {
		t.Fatalf("cast ballot = %v", b)
	}

	deadline := time.Now().Add(2 * time.Second)
	for srv.board.Load().Voters != 1 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if top := srv.board.Load().Top(); len(top) != 5 || top[0].Title != "Stalker" || top[3].Title != "In the Mood for Love" {
		t.Fatalf("board top = %+v", top)
	}

	// Cross-site posts are refused.
	req := httptest.NewRequest("POST", "/ballot/cast", nil)
	req.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cross-site post: %d", w.Code)
	}
}

func digit(i int) string { return string(rune('0' + i)) }

func TestCompanyNote(t *testing.T) {
	for _, c := range []struct {
		others, place int
		want          string
	}{
		{0, 0, "No one else has picked this yet. A discovery."},
		{0, 1, "No one else has picked this yet. A discovery."},
		{1, 0, "You and one other person have this in their five."},
		{37, 12, "You and 37 others have this in their five."},
		{37, 1, "Filmstop’s number one too, with you and 37 others."},
		{4, 3, "Number 3 on Filmstop, with you and 4 others."},
	} {
		if got := companyNote(c.others, c.place); got != c.want {
			t.Errorf("companyNote(%d, %d) = %q, want %q", c.others, c.place, got, c.want)
		}
	}
}
