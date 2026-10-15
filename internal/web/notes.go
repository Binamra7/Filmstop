package web

import (
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/Binamra7/filmstop/internal/store"
)

// Phrase is a run of a note's text. With a MovieID it's a film title, linked to
// the film's page.
type Phrase struct {
	Text    string
	MovieID int64
}

// Note is one line of the programme notes: a short heading and a sentence or two.
type Note struct {
	Head string
	Body []Phrase
}

// Arrival is one recently cast ballot, told as "someone put X first".
type Arrival struct {
	Body []Phrase
	At   time.Time
}

// Thresholds below which a note would be noise rather than news.
const (
	notesMinVoters     = 5
	maxNotes           = 5
	divisiveMinBallots = 6
	divisiveMinSD      = 1.25
	cultMinBallots     = 3
	finishMaxGap       = 3
	finishWithin       = 10
	decadeMinFilms     = 12
	spanMinFilms       = 10
	spanMinYears       = 20
)

// programmeNotes picks the few things about the board worth saying out loud.
// Each note only appears when the numbers back it up, so a young list says
// little and a busy one says more.
func programmeNotes(entries []Entry, voters int) []Note {
	if voters < notesMinVoters || len(entries) < 2 {
		return nil
	}
	var notes []Note
	add := func(n *Note) {
		if n != nil && len(notes) < maxNotes {
			notes = append(notes, *n)
		}
	}
	add(photoFinish(entries))
	divisive := divisiveNote(entries)
	cult := cultNote(entries, divisive)
	add(cult)
	if divisive != nil {
		add(divisive.note)
	}
	add(directorNote(entries))
	add(decadeNote(entries))
	add(spanNote(entries))
	return notes
}

func titled(e Entry) Phrase { return Phrase{Text: e.Title, MovieID: e.MovieID} }
func said(s string) Phrase  { return Phrase{Text: s} }

func photoFinish(es []Entry) *Note {
	best := -1
	for i := 0; i+1 < min(finishWithin, len(es)); i++ {
		gap := es[i].Points - es[i+1].Points
		if gap <= finishMaxGap {
			best = i // the highest-placed close call is the one worth telling
			break
		}
	}
	if best < 0 {
		return nil
	}
	a, b := es[best], es[best+1]
	gap := a.Points - b.Points
	if gap == 0 {
		why := "the tiebreak"
		switch {
		case a.Firsts != b.Firsts:
			why = "first-place votes"
		case a.Ballots != b.Ballots:
			why = "the number of ballots it’s on"
		}
		return &Note{Head: "Photo finish", Body: []Phrase{
			titled(a), said(" and "), titled(b), said(" are level on " + spell(a.Points) + " points. " + a.Title + " takes number " + spell(a.Place) + " on " + why + "."),
		}}
	}
	return &Note{Head: "Photo finish", Body: []Phrase{
		said("Just " + spell(gap) + " " + pluralize(gap, "point separates ", "points separate ")), titled(a),
		said(" at number " + spell(a.Place) + " from "), titled(b),
		said(" at number " + spell(b.Place) + ". One more ballot could swap them."),
	}}
}

type notePick struct {
	id   int64
	note *Note
}

// divisiveNote finds the film voters place most unevenly: loved at the top of
// some ballots, an afterthought at the bottom of others.
func divisiveNote(es []Entry) *notePick {
	var best *Entry
	bestSD := 0.0
	for i := range es {
		e := &es[i]
		if e.Ballots < divisiveMinBallots {
			continue
		}
		top, bottom := e.ByPos[0]+e.ByPos[1], e.ByPos[3]+e.ByPos[4]
		if top < 2 || bottom < 2 {
			continue
		}
		var ss float64
		for p, n := range e.ByPos {
			d := float64(p+1) - e.AvgPos
			ss += float64(n) * d * d
		}
		if sd := math.Sqrt(ss / float64(e.Ballots)); sd >= divisiveMinSD && sd > bestSD {
			best, bestSD = e, sd
		}
	}
	if best == nil {
		return nil
	}
	e := *best
	return &notePick{e.MovieID, &Note{Head: "Split down the middle", Body: []Phrase{
		said("Nobody quite agrees about "), titled(e),
		said(". " + capitalize(spell(e.ByPos[0]+e.ByPos[1])) + " of its " + spell(e.Ballots) + " ballots put it first or second, " +
			spell(e.ByPos[3]+e.ByPos[4]) + " put it fourth or fifth."),
	}}}
}

// cultNote finds the film that, more than any other, is someone's number one
// or nothing. It skips the overall number one, which already has the stage, and
// the divisive pick, so two notes don't describe one film.
func cultNote(es []Entry, divisive *notePick) *Note {
	ballots := make([]int, len(es))
	for i, e := range es {
		ballots[i] = e.Ballots
	}
	sort.Ints(ballots)
	median := ballots[len(ballots)/2]

	var best *Entry
	for i := range es {
		e := &es[i]
		if e.Place == 1 || e.Ballots < cultMinBallots || e.Firsts*2 < e.Ballots {
			continue
		}
		if divisive != nil && divisive.id == e.MovieID {
			continue
		}
		// Higher share of firsts wins; on equal shares, more ballots.
		if best == nil || e.Firsts*best.Ballots > best.Firsts*e.Ballots ||
			(e.Firsts*best.Ballots == best.Firsts*e.Ballots && e.Ballots > best.Ballots) {
			best = e
		}
	}
	if best == nil {
		return nil
	}
	e := *best
	share := capitalize(spell(e.Firsts)) + " of its " + spell(e.Ballots) + " ballots put it at number one."
	if e.Firsts == e.Ballots {
		share = "All " + spell(e.Ballots) + " ballots that list it put it at number one."
	}
	share += " The people who love it love it most."
	if e.Ballots < median {
		return &Note{Head: "A cult favourite", Body: []Phrase{
			said("Fewer people chose "), titled(e), said(" than most films here, but it’s a first choice or nothing. " + share),
		}}
	}
	return &Note{Head: "A cult favourite", Body: []Phrase{
		titled(e), said(" is a first choice more than any film on the list. " + share),
	}}
}

// directorNote names the director with more films on the list than anyone
// else. A shared lead says nothing, so it needs a clear winner.
func directorNote(es []Entry) *Note {
	films := map[string][]Entry{}
	for _, e := range es {
		if e.Director != "" {
			films[e.Director] = append(films[e.Director], e)
		}
	}
	var who string
	most, tied := 0, false
	for d, fs := range films {
		switch {
		case len(fs) > most:
			who, most, tied = d, len(fs), false
		case len(fs) == most:
			tied = true
		}
	}
	if most < 2 || tied {
		return nil
	}
	body := []Phrase{said(who + " has more films on the list than any other director: ")}
	fs := films[who] // already in place order
	for i, e := range fs {
		switch {
		case i == 0:
		case i == len(fs)-1:
			body = append(body, said(" and "))
		default:
			body = append(body, said(", "))
		}
		body = append(body, titled(e))
	}
	body = append(body, said("."))
	return &Note{Head: "Top billing", Body: body}
}

// releaseYear reads a film's year, or 0 if it has none worth trusting. Entry
// embeds store.Movie, so pass e.Movie for a board entry.
func releaseYear(m store.Movie) int {
	if len(m.Year) < 4 {
		return 0
	}
	y, err := strconv.Atoi(m.Year[:4])
	if err != nil || y < 1870 {
		return 0
	}
	return y
}

// decadeNote names the decade with the most films on the list, led by its
// highest-placed film.
func decadeNote(es []Entry) *Note {
	n := 0
	byDecade := map[int][]Entry{}
	for _, e := range es {
		if y := releaseYear(e.Movie); y != 0 {
			n++
			byDecade[y/10*10] = append(byDecade[y/10*10], e)
		}
	}
	if n < decadeMinFilms {
		return nil
	}
	decade, most, tied := 0, 0, false
	for d, fs := range byDecade {
		switch {
		case len(fs) > most:
			decade, most, tied = d, len(fs), false
		case len(fs) == most:
			tied = true
		}
	}
	if tied || most < 4 || most*4 < n {
		return nil
	}
	lead := byDecade[decade][0]
	return &Note{Head: "The list’s decade", Body: []Phrase{
		said("The " + strconv.Itoa(decade) + "s gave us " + spell(most) + " of these " + spell(n) +
			" films, more than any other decade, led by "), titled(lead), said(" at number " + spell(lead.Place) + "."),
	}}
}

// spanNote measures how far back, and how recent, the list reaches.
func spanNote(es []Entry) *Note {
	var old, young Entry
	n := 0
	for _, e := range es {
		y := releaseYear(e.Movie)
		if y == 0 {
			continue
		}
		n++
		if old.Year == "" || y < releaseYear(old.Movie) {
			old = e
		}
		if young.Year == "" || y > releaseYear(young.Movie) {
			young = e
		}
	}
	span := releaseYear(young.Movie) - releaseYear(old.Movie)
	if n < spanMinFilms || span < spanMinYears {
		return nil
	}
	return &Note{Head: strconv.Itoa(span) + " years apart", Body: []Phrase{
		said("The oldest film here is "), titled(old), said(" (" + strconv.Itoa(releaseYear(old.Movie)) + "), and the newest is "),
		titled(young), said(" (" + strconv.Itoa(releaseYear(young.Movie)) + "). Good taste, it turns out, doesn’t keep to an era."),
	}}
}

// arrivals tells the latest ballots as sentences. The wording varies with the
// cast time and film, so a line reads the same on every visit.
func arrivals(os []store.Opener) []Arrival {
	out := make([]Arrival, len(os))
	for i, o := range os {
		f := Phrase{Text: o.Title, MovieID: o.MovieID}
		var body []Phrase
		switch (o.At.Unix() + o.MovieID) % 4 {
		case 0:
			body = []Phrase{said("Someone put "), f, said(" at number one")}
		case 1:
			body = []Phrase{said("Someone opened their five with "), f}
		case 2:
			body = []Phrase{said("A new ballot, led by "), f}
		default:
			body = []Phrase{said("Someone’s favourite film of all time: "), f}
		}
		out[i] = Arrival{Body: body, At: o.At}
	}
	return out
}

var smallNumbers = []string{"no", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve"}

// count spells small numbers out, as a sentence would.
func spell(n int) string {
	if n >= 0 && n < len(smallNumbers) {
		return smallNumbers[n]
	}
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func capitalize(s string) string {
	if s == "" || s[0] < 'a' || s[0] > 'z' {
		return s
	}
	return string(s[0]-'a'+'A') + s[1:]
}

func pluralize(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
