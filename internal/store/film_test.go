package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestFilmTallyAndCompanions(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	var ms []Movie
	for i := int64(1); i <= 9; i++ {
		ms = append(ms, Movie{ID: i, Title: "M"})
	}
	if err := s.UpsertMovies(ctx, ms); err != nil {
		t.Fatal(err)
	}
	for voter, b := range map[string]Ballot{
		"a": {1, 2, 3, 4, 5},
		"b": {2, 1, 3, 6, 7},
		"c": {3, 6, 1, 8, 2},
		"d": {9, 8, 7, 6, 5}, // never meets film 1
	} {
		if err := s.Cast(ctx, voter, b); err != nil {
			t.Fatal(err)
		}
	}

	tl, err := s.FilmTally(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if tl.ByPos != [5]int{1, 1, 1, 0, 0} {
		t.Fatalf("tally = %v", tl.ByPos)
	}
	if tl, _ := s.FilmTally(ctx, 404); tl.ByPos != [5]int{} {
		t.Fatalf("tally of an unpicked film = %v", tl.ByPos)
	}

	got, err := s.Companions(ctx, 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	// 3 and 2 share three ballots each; 3 is ranked higher there (3rd, 3rd, 1st
	// = 11 points against 2nd, 1st, 5th = 10). Then 6 on two, then 4 on one: it
	// ties 8 on points (both 4th) and wins on id, and beats 5 and 7 (5th).
	want := []struct {
		id     int64
		shared int
	}{{3, 3}, {2, 3}, {6, 2}, {4, 1}}
	if len(got) != len(want) {
		t.Fatalf("companions = %+v", got)
	}
	for i, w := range want {
		if got[i].ID != w.id || got[i].Shared != w.shared {
			t.Fatalf("companion %d = %d on %d ballots, want %d on %d", i, got[i].ID, got[i].Shared, w.id, w.shared)
		}
	}
	if none, _ := s.Companions(ctx, 404, 6); len(none) != 0 {
		t.Fatalf("companions of an unpicked film = %+v", none)
	}
}
