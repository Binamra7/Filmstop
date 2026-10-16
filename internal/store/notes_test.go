package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLatestOpeners(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var ms []Movie
	for i := int64(1); i <= 6; i++ {
		ms = append(ms, Movie{ID: i, Title: string(rune('A' + i - 1))})
	}
	if err := s.UpsertMovies(ctx, ms); err != nil {
		t.Fatal(err)
	}
	if got, err := s.LatestOpeners(ctx, 5); err != nil || len(got) != 0 {
		t.Fatalf("empty store: %v, %v", got, err)
	}
	s.Cast(ctx, "a", Ballot{1, 2, 3, 4, 5})
	s.Cast(ctx, "b", Ballot{6, 2, 3, 4, 5})
	s.Cast(ctx, "c", Ballot{3, 2, 1, 4, 5})
	// Spread the cast times so the order is unambiguous.
	for i, v := range []string{"a", "b", "c"} {
		s.db.ExecContext(ctx, `UPDATE ballot_times SET cast_at = ? WHERE voter = ?`, 1000+i, v)
	}
	got, err := s.LatestOpeners(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].MovieID != 3 || got[0].Title != "C" || got[1].MovieID != 6 {
		t.Fatalf("latest = %+v", got)
	}
	if got[0].At.Unix() != 1002 {
		t.Fatalf("at = %v", got[0].At)
	}
}
