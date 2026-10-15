package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOthersWith(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var ms []Movie
	for i := int64(1); i <= 8; i++ {
		ms = append(ms, Movie{ID: i, Title: "M"})
	}
	if err := s.UpsertMovies(ctx, ms); err != nil {
		t.Fatal(err)
	}
	for voter, b := range map[string]Ballot{
		"a": {1, 2, 3, 4, 5},
		"b": {5, 1, 2, 6, 7},
		"c": {1, 6, 2, 3, 7},
	} {
		if err := s.Cast(ctx, voter, b); err != nil {
			t.Fatal(err)
		}
	}
	// A draft isn't a counted ballot.
	if err := s.SaveDraft(ctx, "d", Ballot{8, 0, 0, 0, 0}); err != nil {
		t.Fatal(err)
	}

	got, err := s.OthersWith(ctx, "a", []int64{1, 2, 4, 7, 8})
	if err != nil {
		t.Fatal(err)
	}
	want := map[int64]int{1: 2, 2: 2, 7: 2} // 4 is only on a's own ballot, 8 only in a draft
	if len(got) != len(want) {
		t.Fatalf("company = %v, want %v", got, want)
	}
	for id, n := range want {
		if got[id] != n {
			t.Fatalf("company = %v, want %v", got, want)
		}
	}
	if got, err := s.OthersWith(ctx, "nobody", nil); err != nil || len(got) != 0 {
		t.Fatalf("empty ids: %v %v", got, err)
	}
}
