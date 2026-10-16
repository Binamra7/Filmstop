package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Binamra7/filmstop/internal/rank"
	"github.com/Binamra7/filmstop/internal/store"
	"github.com/Binamra7/filmstop/internal/tmdb"
	"github.com/gin-gonic/gin"
)

func TestFilmPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	titles := []string{"Stalker", "Yi Yi", "Heat", "Paris, Texas", "Ran", "Amélie"}
	var ms []store.Movie
	for i, title := range titles {
		ms = append(ms, store.Movie{ID: int64(i + 1), Title: title})
	}
	st.UpsertMovies(ctx, ms)
	st.Cast(ctx, "a", store.Ballot{1, 2, 3, 4, 5})
	srv, err := New(ctx, st, tmdb.New(""))
	if err != nil {
		t.Fatal(err)
	}
	h := srv.Handler()
	get := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		return w
	}

	w := get("/film/1")
	body := w.Body.String()
	if w.Code != http.StatusOK || !strings.Contains(body, "No. 1 on Filmstop") || !strings.Contains(body, `href="/film/2"`) {
		t.Fatalf("listed film: %d\n%s", w.Code, body)
	}
	if !strings.Contains(body, "they put it first") || !strings.Contains(body, `action="/ballot/pick"`) {
		t.Fatalf("listed film is missing its note or its pick form:\n%s", body)
	}
	if w := get("/film/6"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "No one has picked it yet") {
		t.Fatalf("unpicked film: %d", w.Code)
	}
	for _, p := range []string{"/film/999", "/film/abc", "/film/-1"} {
		if w := get(p); w.Code != http.StatusNotFound {
			t.Fatalf("%s: %d", p, w.Code)
		}
	}
}

func TestShortOfList(t *testing.T) {
	last := Entry{Standing: rank.Standing{Place: 30, Points: 12}, Movie: store.Movie{Title: "Ran"}}
	for pts, want := range map[int]string{
		11: "1 point behind Ran, number 30. One first place",
		5:  "7 points behind Ran, number 30. That’s 2 first places away.",
		12: "level with Ran",
	} {
		if got := shortOfList(rank.Standing{Points: pts}, last); !strings.Contains(got, want) {
			t.Errorf("%d points: %q, want %q", pts, got, want)
		}
	}
}
