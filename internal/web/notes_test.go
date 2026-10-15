package web

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Binamra7/filmstop/internal/rank"
	"github.com/Binamra7/filmstop/internal/store"
)

// board ranks tallies the way the keeper does and dresses them as entries.
func board(ts []rank.Tally, movie func(id int64) store.Movie) []Entry {
	var es []Entry
	for _, s := range rank.Order(ts) {
		es = append(es, Entry{Standing: s, Movie: movie(s.MovieID)})
	}
	return es
}

func plain(n Note) string {
	var b strings.Builder
	for _, p := range n.Body {
		if p.MovieID != 0 {
			fmt.Fprintf(&b, "[%s]", p.Text)
		} else {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

func heads(ns []Note) map[string]string {
	out := map[string]string{}
	for _, n := range ns {
		out[n.Head] = plain(n)
	}
	return out
}

func TestProgrammeNotesQuietOnSmallLists(t *testing.T) {
	es := board([]rank.Tally{
		{MovieID: 1, ByPos: [5]int{2, 0, 0, 0, 0}},
		{MovieID: 2, ByPos: [5]int{0, 2, 0, 0, 0}},
	}, func(id int64) store.Movie { return store.Movie{ID: id, Title: "F", Year: "2000"} })
	if ns := programmeNotes(es, 2); len(ns) != 0 {
		t.Fatalf("two voters should get no notes, got %v", ns)
	}
	if ns := programmeNotes(nil, 100); len(ns) != 0 {
		t.Fatalf("empty board should get no notes, got %v", ns)
	}
}

func TestProgrammeNotes(t *testing.T) {
	ts := []rank.Tally{
		{MovieID: 1, ByPos: [5]int{10, 5, 0, 0, 0}}, // 65 points, the number one
		{MovieID: 2, ByPos: [5]int{4, 1, 0, 1, 6}},  // split: 4+1 high, 1+6 low
		{MovieID: 3, ByPos: [5]int{0, 4, 3, 2, 0}},  // 29, three behind film 2
		{MovieID: 4, ByPos: [5]int{0, 4, 3, 1, 1}},  // 28
		{MovieID: 5, ByPos: [5]int{3, 0, 0, 0, 0}},  // all firsts, few ballots
		{MovieID: 6, ByPos: [5]int{0, 0, 2, 2, 2}},
	}
	for id := int64(7); id <= 16; id++ {
		ts = append(ts, rank.Tally{MovieID: id, ByPos: [5]int{0, 0, 0, 1, 1}})
	}
	years := map[int64]string{1: "1927-03-01", 2: "1994", 3: "1995", 4: "1999", 5: "2023", 6: "1991"}
	es := board(ts, func(id int64) store.Movie {
		m := store.Movie{ID: id, Title: fmt.Sprintf("Film %d", id), Year: years[id]}
		if m.Year == "" {
			m.Year = fmt.Sprint(1950 + id*4)
		}
		if id == 3 || id == 6 || id == 9 {
			m.Director = "Tsai Ming-liang"
		}
		if id == 4 {
			m.Director = "Edward Yang"
		}
		return m
	})
	ns := programmeNotes(es, 40)
	if len(ns) > maxNotes {
		t.Fatalf("%d notes, want at most %d", len(ns), maxNotes)
	}
	got := heads(ns)
	want := map[string]string{
		"Photo finish":          "Just three points separate [Film 2] at number two from [Film 3] at number three. One more ballot could swap them.",
		"A cult favourite":      "[Film 5] is a first choice more than any film on the list. All three ballots that list it put it at number one.",
		"Split down the middle": "about [Film 2]. Five of its twelve ballots put it first or second, seven put it fourth or fifth.",
		"Top billing":           "Tsai Ming-liang has more films on the list than any other director: [Film 3], [Film 6] and [Film 9].",
		"The list’s decade":     "The 1990s gave us seven of these 16 films",
	}
	for head, frag := range want {
		if !strings.Contains(got[head], frag) {
			t.Errorf("%q = %q, want it to contain %q", head, got[head], frag)
		}
	}
	// The span note would be sixth, past the limit.
	if s, ok := got["96 years apart"]; ok {
		t.Errorf("unexpected span note %q", s)
	}
	if n := spanNote(es); n == nil || !strings.Contains(plain(*n), "oldest film here is [Film 1] (1927), and the newest is [Film 5] (2023)") {
		t.Errorf("span note = %+v", n)
	}
}

func TestDecadeNote(t *testing.T) {
	var ts []rank.Tally
	for id := int64(1); id <= 16; id++ {
		ts = append(ts, rank.Tally{MovieID: id, ByPos: [5]int{int(20 - id), 0, 0, 0, 0}})
	}
	es := board(ts, func(id int64) store.Movie {
		y := 1960 + int(id)*3 // 1963 .. 2008
		if id%3 == 0 {
			y = 1990 + int(id)%10
		}
		return store.Movie{ID: id, Title: fmt.Sprintf("Film %d", id), Year: fmt.Sprint(y)}
	})
	n := decadeNote(es)
	if n == nil {
		t.Fatal("want a decade note")
	}
	if s := plain(*n); !strings.HasPrefix(s, "The 1990s gave us eight of these 16 films") || !strings.Contains(s, "[Film 3] at number three") {
		t.Fatalf("decade note = %q", s)
	}
}

func TestDirectorNoteNeedsAClearLead(t *testing.T) {
	es := []Entry{
		{Movie: store.Movie{Title: "A", Details: store.Details{Director: "X"}}},
		{Movie: store.Movie{Title: "B", Details: store.Details{Director: "X"}}},
		{Movie: store.Movie{Title: "C", Details: store.Details{Director: "Y"}}},
		{Movie: store.Movie{Title: "D", Details: store.Details{Director: "Y"}}},
		{Movie: store.Movie{Title: "E"}},
	}
	if n := directorNote(es); n != nil {
		t.Fatalf("tied directors should get no note, got %q", plain(*n))
	}
}

func TestPhotoFinishTie(t *testing.T) {
	es := board([]rank.Tally{
		{MovieID: 1, ByPos: [5]int{2, 0, 0, 0, 0}},
		{MovieID: 2, ByPos: [5]int{1, 0, 0, 0, 5}},
	}, func(id int64) store.Movie { return store.Movie{ID: id, Title: fmt.Sprintf("Film %d", id)} })
	n := photoFinish(es)
	if n == nil || plain(*n) != "[Film 1] and [Film 2] are level on ten points. Film 1 takes number one on first-place votes." {
		t.Fatalf("photo finish = %+v", n)
	}
}

func TestArrivalsNeverNameVoters(t *testing.T) {
	as := arrivals([]store.Opener{{MovieID: 7, Title: "Stalker", At: time.Unix(1000, 0)}})
	if len(as) != 1 || as[0].At.Unix() != 1000 {
		t.Fatalf("arrivals = %+v", as)
	}
	var linked bool
	for _, p := range as[0].Body {
		linked = linked || (p.MovieID == 7 && p.Text == "Stalker")
	}
	if !linked {
		t.Fatalf("arrival should link the film: %+v", as[0].Body)
	}
}
