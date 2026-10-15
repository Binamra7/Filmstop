package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestDraftCastAndTally(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	var ms []Movie
	for i := int64(1); i <= 6; i++ {
		ms = append(ms, Movie{ID: i, Title: "M"})
	}
	if err := s.UpsertMovies(ctx, ms); err != nil {
		t.Fatal(err)
	}

	if err := s.Cast(ctx, "a", Ballot{1, 2, 3, 0, 0}); err == nil {
		t.Fatal("cast of a partial ballot should fail")
	}
	if err := s.SaveDraft(ctx, "a", Ballot{1, 2, 0, 0, 0}); err != nil {
		t.Fatal(err)
	}
	if b, _ := s.Draft(ctx, "a"); b != (Ballot{1, 2, 0, 0, 0}) {
		t.Fatalf("draft = %v", b)
	}
	if err := s.Cast(ctx, "a", Ballot{1, 2, 3, 4, 5}); err != nil {
		t.Fatal(err)
	}
	// Draft is cleared and falls back to the cast ballot.
	if b, _ := s.Draft(ctx, "a"); b != (Ballot{1, 2, 3, 4, 5}) {
		t.Fatalf("draft after cast = %v", b)
	}
	// Recasting replaces, not adds.
	if err := s.Cast(ctx, "a", Ballot{6, 2, 3, 4, 5}); err != nil {
		t.Fatal(err)
	}
	if err := s.Cast(ctx, "b", Ballot{6, 1, 2, 3, 4}); err != nil {
		t.Fatal(err)
	}
	ts, voters, err := s.Tallies(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if voters != 2 {
		t.Fatalf("voters = %d", voters)
	}
	byID := map[int64][5]int{}
	for _, tl := range ts {
		byID[tl.MovieID] = tl.ByPos
	}
	if byID[6] != [5]int{2, 0, 0, 0, 0} || byID[1] != [5]int{0, 1, 0, 0, 0} {
		t.Fatalf("tallies = %v", byID)
	}
	if ok, _ := s.HasCast(ctx, "b"); !ok {
		t.Fatal("b should have cast")
	}
	sec1, _ := s.Secret(ctx)
	sec2, _ := s.Secret(ctx)
	if len(sec1) != 32 || string(sec1) != string(sec2) {
		t.Fatal("secret should be stable 32 bytes")
	}
}

func TestClearedDraftStaysClear(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var ms []Movie
	for i := int64(1); i <= 5; i++ {
		ms = append(ms, Movie{ID: i, Title: "M"})
	}
	if err := s.UpsertMovies(ctx, ms); err != nil {
		t.Fatal(err)
	}
	if err := s.Cast(ctx, "a", Ballot{1, 2, 3, 4, 5}); err != nil {
		t.Fatal(err)
	}
	b, err := s.EditDraft(ctx, "a", func(b *Ballot) { *b = Ballot{} })
	if err != nil || b != (Ballot{}) {
		t.Fatalf("edit = %v, %v", b, err)
	}
	// Removing every scene leaves an empty draft, not the cast ballot.
	if b, _ := s.Draft(ctx, "a"); b != (Ballot{}) {
		t.Fatalf("cleared draft = %v", b)
	}
	// Casting closes the draft, so the next one starts from the ballot again.
	if err := s.Cast(ctx, "a", Ballot{5, 4, 3, 2, 1}); err != nil {
		t.Fatal(err)
	}
	if b, _ := s.Draft(ctx, "a"); b != (Ballot{5, 4, 3, 2, 1}) {
		t.Fatalf("draft after recast = %v", b)
	}
}
