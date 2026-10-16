package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Binamra7/filmstop/internal/store"
	"github.com/Binamra7/filmstop/internal/tmdb"
	"github.com/gin-gonic/gin"
)

func film(id int64, title, year string) store.Movie {
	return store.Movie{ID: id, Title: title, Year: year}
}

func (t taste) text() string {
	var b strings.Builder
	for _, p := range t.Paras {
		b.WriteString(p.String())
		b.WriteString("\n")
	}
	return b.String()
}

func TestReadTaste(t *testing.T) {
	nineties := []store.Movie{
		film(1, "Heat", "1995"), film(2, "Chungking Express", "1994"), film(3, "Stalker", "1979"),
		film(4, "Fargo", "1996"), film(5, "Yi Yi", "2000"),
	}
	cases := []struct {
		name     string
		in       tasteInput
		headline string
		want     []string
		not      []string
	}{
		{
			name: "nineties, partly mainstream, one secret",
			in: tasteInput{
				Owner: true, Picks: nineties, Voters: 40,
				Places:   map[int64]int{1: 1, 2: 7, 4: 22},
				BoardOne: nineties[0],
				Counts:   map[int64]int{1: 20, 2: 6, 3: 4, 4: 3, 5: 1},
				Kindred:  []int64{2, 1, 3},
			},
			headline: "Firmly set in the 1990s.",
			want: []string{
				"You live in the 1990s: three of the five come from that decade.",
				"Oldest is Stalker (1979), newest Yi Yi (2000).",
				"Three of the five are on Filmstop’s list; Stalker and Yi Yi are your own business.",
				"And Heat is Filmstop’s number one too. Great minds.",
				"Nobody else has picked Yi Yi. That one is your secret.",
				"shares three of these: Heat, Chungking Express and Stalker. A kindred spirit.",
			},
			not: []string{"didn’t make the cut"},
		},
		{
			name: "a stranger's out-of-step five",
			in: tasteInput{
				Picks: []store.Movie{
					film(11, "Sátántangó", "1994"), film(12, "Close-Up", "1990"), film(13, "Jeanne Dielman", "1975"),
					film(14, "Wanda", "1970"), film(15, "Sans Soleil", "1983"),
				},
				Voters: 12, Places: map[int64]int{1: 1}, BoardOne: film(1, "Heat", "1995"),
				Counts:  map[int64]int{11: 1, 12: 2, 13: 1, 14: 1, 15: 1},
				Kindred: []int64{12},
			},
			headline: "Out of step, on purpose.",
			want: []string{
				"Not one of these is on Filmstop’s list. Gloriously out of step.",
				"Filmstop’s own number one, Heat, didn’t make the cut.",
				"Nobody else has picked Sátántangó, Jeanne Dielman, Wanda or Sans Soleil. Those are their secrets.",
				"The closest anyone comes is one shared film: Close-Up.",
				"Split down the middle between the 1970s and the 1990s",
			},
			not: []string{"You", "your", "hasn’t caught up"},
		},
		{
			name: "five decades, half a century",
			in: tasteInput{
				Owner: true, Voters: 3,
				Picks: []store.Movie{
					film(21, "Ikiru", "1952"), film(22, "Persona", "1966"), film(23, "Alien", "1979"),
					film(24, "Ran", "1985"), film(25, "Parasite", "2019"),
				},
				Places:  map[int64]int{21: 2, 22: 3, 23: 4, 24: 5, 25: 6},
				Counts:  map[int64]int{21: 2, 22: 2, 23: 2, 24: 2, 25: 2},
				Kindred: []int64{21, 22, 23, 24, 25},
			},
			headline: "Great minds, exactly.",
			want: []string{
				"Five films, five different decades. You don’t do eras.",
				"It runs from Ikiru (1952) to Parasite (2019): 67 years of cinema in five films.",
				"All five are on Filmstop’s list. You’re the consensus",
				"Your number one, Ikiru, sits at No. 2 on the list.",
				"Somebody else picked the very same five.",
			},
			not: []string{"secret"},
		},
		{
			name:     "first ballot in",
			in:       tasteInput{Owner: true, Picks: nineties, Voters: 1, Counts: map[int64]int{1: 1, 2: 1, 3: 1, 4: 1, 5: 1}},
			headline: "The first five in.",
			want:     []string{"This is the first ballot Filmstop has counted"},
			not:      []string{"kindred", "Nobody else"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := readTaste(tc.in)
			text := got.text()
			if got.Headline != tc.headline {
				t.Errorf("headline = %q, want %q", got.Headline, tc.headline)
			}
			for _, w := range tc.want {
				if !strings.Contains(text, w) {
					t.Errorf("missing %q in:\n%s", w, text)
				}
			}
			for _, w := range tc.not {
				if strings.Contains(text, w) {
					t.Errorf("unexpected %q in:\n%s", w, text)
				}
			}
		})
	}
}

func TestFiveRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	var ms []store.Movie
	for i, title := range []string{"Stalker", "Yi Yi", "Heat", "Paris, Texas", "In the Mood for Love"} {
		ms = append(ms, film(int64(i+1), title, "1990"))
	}
	st.UpsertMovies(ctx, ms)
	srv, err := New(ctx, st, tmdb.New(""))
	if err != nil {
		t.Fatal(err)
	}
	h := srv.Handler()
	get := func(path string, cookie *http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", path, nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w
	}

	// No ballot yet: off to the ballot.
	w := get("/five", nil)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/ballot" {
		t.Fatalf("/five without a ballot: %d %s", w.Code, w.Header().Get("Location"))
	}
	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == voterCookie {
			cookie = c
		}
	}
	voter, _ := srv.voters.verify(cookie.Value)
	if err := st.Cast(ctx, voter, store.Ballot{1, 2, 3, 4, 5}); err != nil {
		t.Fatal(err)
	}

	w = get("/five", cookie)
	loc := w.Header().Get("Location")
	if w.Code != http.StatusSeeOther || !strings.HasPrefix(loc, "/five/") || strings.Contains(loc, voter) {
		t.Fatalf("/five after casting: %d %s", w.Code, loc)
	}
	if again := get("/five", cookie).Header().Get("Location"); again != loc {
		t.Fatalf("share code not stable: %s then %s", loc, again)
	}

	own := get(loc, cookie).Body.String()
	if !strings.Contains(own, "Copy link") || !strings.Contains(own, "Edit your five") || !strings.Contains(own, "Paris, Texas") {
		t.Fatal("owner card is missing its actions or films")
	}
	theirs := get(loc, nil)
	if theirs.Code != 200 || !strings.Contains(theirs.Body.String(), "Pick your own five") || strings.Contains(theirs.Body.String(), "Copy link") {
		t.Fatal("visitor card should offer a ballot, not the owner's tools")
	}
	if strings.Contains(theirs.Body.String(), voter) {
		t.Fatal("card leaks the voter id")
	}
	for _, bad := range []string{"/five/zzzzzzzzzz", "/five/short", "/five/" + voter} {
		if w := get(bad, nil); w.Code != http.StatusNotFound {
			t.Fatalf("%s: %d", bad, w.Code)
		}
	}
}
