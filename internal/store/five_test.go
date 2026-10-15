package store

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
)

func TestShareCodesAndCompany(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	code, err := s.ShareCode(ctx, "a")
	if err != nil || !ValidCode(code) {
		t.Fatalf("code = %q, %v", code, err)
	}
	if again, _ := s.ShareCode(ctx, "a"); again != code {
		t.Fatalf("code changed: %q then %q", code, again)
	}
	if other, _ := s.ShareCode(ctx, "b"); other == code {
		t.Fatal("two voters share a code")
	}
	if v, ok, err := s.VoterByCode(ctx, code); err != nil || !ok || v != "a" {
		t.Fatalf("lookup = %q %v %v", v, ok, err)
	}
	if _, ok, _ := s.VoterByCode(ctx, "zzzzzzzzzz"); ok {
		t.Fatal("unknown code found")
	}
	if ValidCode("ABCDEFGHJK") || ValidCode("abc") || ValidCode("abcdefghi0") {
		t.Fatal("ValidCode accepts junk")
	}

	var ms []Movie
	for i := int64(1); i <= 9; i++ {
		ms = append(ms, Movie{ID: i, Title: "M"})
	}
	s.UpsertMovies(ctx, ms)
	for v, b := range map[string]Ballot{
		"a": {1, 2, 3, 4, 5},
		"b": {6, 7, 8, 9, 1}, // shares 1
		"c": {5, 4, 3, 7, 8}, // shares 3, far apart
		"d": {9, 2, 3, 4, 6}, // shares 3, close: the kindred one
		"e": {6, 7, 8, 9, 2}, // shares 2
	} {
		if err := s.Cast(ctx, v, b); err != nil {
			t.Fatal(err)
		}
	}
	counts, kin, err := s.Company(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	want := map[int64]int{1: 2, 2: 3, 3: 3, 4: 3, 5: 2}
	for id, n := range want {
		if counts[id] != n {
			t.Fatalf("counts = %v, want %v", counts, want)
		}
	}
	slices.Sort(kin.Shared)
	if !slices.Equal(kin.Shared, []int64{2, 3, 4}) {
		t.Fatalf("kindred shared = %v", kin.Shared)
	}
	if b, _ := s.CastBallot(ctx, "a"); b != (Ballot{1, 2, 3, 4, 5}) {
		t.Fatalf("cast ballot = %v", b)
	}
	if _, kin, _ := s.Company(ctx, "nobody"); kin.Shared != nil {
		t.Fatal("no ballot should have no kindred")
	}
}
