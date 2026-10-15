// Package rank turns many people's top-five ballots into one ordering.
//
// Every ballot is a top-5 list drawn from all films ever made, so any model that
// treats unlisted films as "ranked below" (Plackett–Luce, Condorcet/Schulze) learns
// almost everything from whether a film was listed and almost nothing from where:
// a film always ranked #1 and one always ranked #5 by the same number of people come
// out within a few percent of each other. Filmstop's premise is that position
// matters, so we use a positional score instead and make the weights explicit.
// See the method page for the full reasoning.
package rank

import "sort"

// Slots is the length of a ballot.
const Slots = 5

// Points is what a film earns for appearing at a position (1-based) on one ballot.
// A first place is worth five fifth places.
func Points(position int) int {
	if position < 1 || position > Slots {
		return 0
	}
	return Slots + 1 - position
}

// Tally is one film's raw counts: ByPos[0] is how many ballots ranked it #1, etc.
type Tally struct {
	MovieID int64
	ByPos   [Slots]int
}

// Standing is a film's place in the overall list.
type Standing struct {
	MovieID int64
	Place   int
	Points  int
	Ballots int
	Firsts  int
	// AvgPos is the mean position among ballots that listed it.
	AvgPos float64
	ByPos  [Slots]int
}

// Order scores tallies and sorts them. Ties on points go to the film with more
// first places, then the one listed on more ballots, then the lower id (older
// TMDB entry), so the order is always deterministic.
func Order(tallies []Tally) []Standing {
	out := make([]Standing, 0, len(tallies))
	for _, t := range tallies {
		s := Standing{MovieID: t.MovieID, Firsts: t.ByPos[0], ByPos: t.ByPos}
		posSum := 0
		for i, n := range t.ByPos {
			s.Points += n * Points(i+1)
			s.Ballots += n
			posSum += n * (i + 1)
		}
		if s.Ballots == 0 {
			continue
		}
		s.AvgPos = float64(posSum) / float64(s.Ballots)
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Points != b.Points {
			return a.Points > b.Points
		}
		if a.Firsts != b.Firsts {
			return a.Firsts > b.Firsts
		}
		if a.Ballots != b.Ballots {
			return a.Ballots > b.Ballots
		}
		return a.MovieID < b.MovieID
	})
	for i := range out {
		out[i].Place = i + 1
	}
	return out
}
