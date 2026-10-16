package rank

import (
	"math"
	"testing"
)

func TestPoints(t *testing.T) {
	want := map[int]int{0: 0, 1: 5, 2: 4, 3: 3, 4: 2, 5: 1, 6: 0}
	for pos, p := range want {
		if got := Points(pos); got != p {
			t.Errorf("Points(%d) = %d, want %d", pos, got, p)
		}
	}
}

func TestOrderPositionMatters(t *testing.T) {
	// Same number of ballots; one film is always first, the other always fifth.
	got := Order([]Tally{
		{MovieID: 2, ByPos: [Slots]int{0, 0, 0, 0, 10}},
		{MovieID: 1, ByPos: [Slots]int{10, 0, 0, 0, 0}},
	})
	if got[0].MovieID != 1 || got[0].Points != 50 || got[1].Points != 10 {
		t.Fatalf("got %+v", got)
	}
}

func TestOrderBreadthCanBeatDepth(t *testing.T) {
	// 100 fifth places outweigh 19 first places.
	got := Order([]Tally{
		{MovieID: 1, ByPos: [Slots]int{19, 0, 0, 0, 0}},
		{MovieID: 2, ByPos: [Slots]int{0, 0, 0, 0, 100}},
	})
	if got[0].MovieID != 2 {
		t.Fatalf("got %+v", got)
	}
}

func TestOrderTieBreaks(t *testing.T) {
	got := Order([]Tally{
		{MovieID: 5, ByPos: [Slots]int{0, 0, 0, 0, 0}}, // no ballots: dropped
		{MovieID: 4, ByPos: [Slots]int{0, 1, 0, 0, 1}}, // 5 pts, 0 firsts
		{MovieID: 3, ByPos: [Slots]int{1, 0, 0, 0, 0}}, // 5 pts, 1 first
		{MovieID: 2, ByPos: [Slots]int{0, 0, 1, 1, 0}}, // 5 pts, 0 firsts, 2 ballots, lower id
	})
	ids := []int64{3, 2, 4}
	if len(got) != len(ids) {
		t.Fatalf("got %d standings, want %d", len(got), len(ids))
	}
	for i, id := range ids {
		if got[i].MovieID != id || got[i].Place != i+1 {
			t.Fatalf("place %d: got %+v", i+1, got[i])
		}
	}
	if math.Abs(got[1].AvgPos-3.5) > 1e-9 {
		t.Errorf("AvgPos = %v, want 3.5", got[1].AvgPos)
	}
}
