package rank

import "testing"

// fitPlackettLuce fits Plackett–Luce weights to top-k ballots by Hunter's (2004)
// MM algorithm, with every unlisted film still in the choice set. It exists only to
// back the claim in the package doc; production ranking does not use it.
func fitPlackettLuce(ballots [][]int, films int, iters int) []float64 {
	w := make([]float64, films)
	for i := range w {
		w[i] = 1
	}
	wins := make([]float64, films)
	for _, b := range ballots {
		for _, f := range b {
			wins[f]++
		}
	}
	for range iters {
		total := 0.0
		for _, x := range w {
			total += x
		}
		denom := make([]float64, films)
		global := 0.0
		for _, b := range ballots {
			// prefix[j] = sum of 1/S over stages 0..j; film b[j] is in stages 0..j.
			s, acc := total, 0.0
			prefix := make([]float64, len(b))
			for j, f := range b {
				acc += 1 / s
				prefix[j] = acc
				s -= w[f]
			}
			global += acc
			for j, f := range b {
				denom[f] -= acc - prefix[j]
			}
		}
		for i := range w {
			w[i] = wins[i] / (global + denom[i])
		}
	}
	return w
}

func TestPlackettLuceIgnoresPosition(t *testing.T) {
	// 500 films, 1000 voters. Film 0 is everyone-in-group-A's #1; film 1 is
	// everyone-in-group-B's #5. Same breadth, opposite depth.
	const films, voters = 500, 1000
	var ballots [][]int
	for v := range voters {
		filler := func(k int) int { return 2 + (v*7+k*13)%(films-2) }
		if v%10 == 0 {
			ballots = append(ballots, []int{0, filler(1), filler(2), filler(3), filler(4)})
		} else if v%10 == 5 {
			ballots = append(ballots, []int{filler(1), filler(2), filler(3), filler(4), 1})
		} else {
			ballots = append(ballots, []int{filler(0), filler(1), filler(2), filler(3), filler(4)})
		}
	}
	w := fitPlackettLuce(ballots, films, 200)
	ratio := w[0] / w[1]
	t.Logf("Plackett–Luce: always-#1 / always-#5 strength = %.3f", ratio)
	if ratio > 1.15 {
		t.Fatalf("expected position to barely matter under PL, ratio %.3f", ratio)
	}

	tallies := []Tally{{MovieID: 0, ByPos: [Slots]int{100}}, {MovieID: 1, ByPos: [Slots]int{4: 100}}}
	got := Order(tallies)
	t.Logf("Positional points: always-#1 = %d, always-#5 = %d", got[0].Points, got[1].Points)
}
