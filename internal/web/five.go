package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Binamra7/filmstop/internal/store"
	"github.com/gin-gonic/gin"
)

// fiveRoutes registers the shareable "your five" card routes.
func (s *Server) fiveRoutes(r gin.IRoutes) {
	r.GET("/five", s.myFive)
	r.GET("/five/:code", s.fiveCard)
}

// myFive sends a voter to their own card, or to the ballot if they haven't cast.
func (s *Server) myFive(c *gin.Context) {
	voter := voterID(c)
	cast, err := s.st.HasCast(c.Request.Context(), voter)
	if err != nil {
		s.fail(c, err)
		return
	}
	if !cast {
		c.Redirect(http.StatusSeeOther, "/ballot")
		return
	}
	code, err := s.st.ShareCode(c.Request.Context(), voter)
	if err != nil {
		s.fail(c, err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/five/"+code)
}

type fivePick struct {
	Pos   int
	Place int // on Filmstop's list, 0 if it isn't there
	store.Movie
}

type fiveView struct {
	page
	Code  string
	Owner bool
	One   fivePick
	Rest  []fivePick
	Taste taste
}

func (s *Server) fiveCard(c *gin.Context) {
	code := c.Param("code")
	notFound := func() { s.render(c, http.StatusNotFound, "notfound", page{"Not found · Filmstop", ""}) }
	if !store.ValidCode(code) {
		notFound()
		return
	}
	voter, ok, err := s.st.VoterByCode(c.Request.Context(), code)
	if err != nil {
		s.fail(c, err)
		return
	}
	if !ok {
		notFound()
		return
	}
	b, err := s.st.CastBallot(c.Request.Context(), voter)
	if err != nil {
		s.fail(c, err)
		return
	}
	if !b.Full() {
		notFound()
		return
	}
	movies, err := s.st.Movies(c.Request.Context(), b[:])
	if err != nil {
		s.fail(c, err)
		return
	}
	counts, kin, err := s.st.Company(c.Request.Context(), voter)
	if err != nil {
		s.fail(c, err)
		return
	}
	// The board can lag a fresh cast, so count the voters now: a card opened
	// straight after the second ballot must not call itself the first.
	nVoters, err := s.st.Voters(c.Request.Context())
	if err != nil {
		s.fail(c, err)
		return
	}
	board := s.board.Load()
	places := make(map[int64]int, len(board.Entries))
	for _, e := range board.Entries {
		places[e.MovieID] = e.Place
	}

	owner := voterID(c) == voter
	in := tasteInput{Owner: owner, Places: places, Counts: counts, Voters: nVoters, Kindred: kin.Shared}
	if len(board.Entries) > 0 {
		in.BoardOne = board.Entries[0].Movie
	}
	var picks []fivePick
	for i, id := range b {
		m := movies[id]
		in.Picks = append(in.Picks, m)
		picks = append(picks, fivePick{Pos: i + 1, Place: places[id], Movie: m})
	}
	s.fiveDetails(in.Picks)

	title := picks[0].Title + " and four more · Filmstop"
	if owner {
		title = "Your five · Filmstop"
	}
	s.render(c, http.StatusOK, "five", fiveView{
		page: page{title, "ballot"}, Code: code, Owner: owner,
		One: picks[0], Rest: picks[1:], Taste: readTaste(in),
	})
}

// fiveDetailing holds film ids whose details are being fetched for a card.
var fiveDetailing sync.Map

// fiveDetails fetches directors and stills, in the background, for picks that
// aren't on the board (the board keeper only details films on the list). The
// next view of the card shows them.
func (s *Server) fiveDetails(ms []store.Movie) {
	if !s.tmdb.Enabled() {
		return
	}
	for _, m := range ms {
		if m.Detailed {
			continue
		}
		if _, busy := fiveDetailing.LoadOrStore(m.ID, true); busy {
			continue
		}
		go func(id int64) {
			defer fiveDetailing.Delete(id)
			ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
			defer cancel()
			backdrop, poster, d, err := s.tmdb.Details(ctx, id)
			if err == nil {
				err = s.st.SetDetails(ctx, id, backdrop, poster, d)
			}
			if err != nil {
				slog.Warn("film details", "id", id, "err", err)
			}
		}(m.ID)
	}
}

// --- the taste reading

type tasteInput struct {
	Owner    bool          // written to "you"; otherwise about "them"
	Picks    []store.Movie // the ballot, best first
	Places   map[int64]int // film -> place on Filmstop's list (top 30)
	BoardOne store.Movie   // Filmstop's number one (zero if there's no list)
	Counts   map[int64]int // film -> ballots that list it, this one included
	Voters   int           // ballots counted
	Kindred  []int64       // films shared with the closest other ballot
}

// A taste reading is a headline and a few paragraphs. Paragraphs are runs of
// plain text and film titles, so the template can set titles in the title face.
type taste struct {
	Headline string
	Paras    []tastePara
}

type tastePara []tasteRun

type tasteRun struct {
	Text string
	Film bool
}

// String returns the paragraph as plain text, titles included.
func (p tastePara) String() string {
	var b strings.Builder
	for _, r := range p {
		b.WriteString(r.Text)
	}
	return b.String()
}

// tasteVoice holds the pronouns, so one reading works on your card and a friend's.
type tasteVoice struct{ S, O, P, Re string }

var (
	voiceYou  = tasteVoice{"You", "you", "your", "you’re"}
	voiceThem = tasteVoice{"They", "them", "their", "they’re"}
)

// tasteLine builds a paragraph from text and titles.
type tasteLine struct{ p tastePara }

func (b *tasteLine) t(format string, a ...any) *tasteLine {
	b.p = append(b.p, tasteRun{Text: fmt.Sprintf(format, a...)})
	return b
}

func (b *tasteLine) film(m store.Movie) *tasteLine {
	b.p = append(b.p, tasteRun{Text: m.Title, Film: true})
	return b
}

// films writes "A, B and C" (or "or" for a negative list).
func (b *tasteLine) films(ms []store.Movie, conj string) *tasteLine {
	for i, m := range ms {
		switch {
		case i == 0:
		case i == len(ms)-1:
			b.t(" %s ", conj)
		default:
			b.t(", ")
		}
		b.film(m)
	}
	return b
}

func readTaste(in tasteInput) taste {
	v := voiceThem
	if in.Owner {
		v = voiceYou
	}
	var t taste
	add := func(b *tasteLine) {
		if len(b.p) > 0 {
			t.Paras = append(t.Paras, b.p)
		}
	}

	// When it was made: the decade it leans on, and the span.
	decades := map[int]int{}
	var dated []store.Movie
	for _, m := range in.Picks {
		if y := releaseYear(m); y > 0 {
			decades[y/10*10]++
			dated = append(dated, m)
		}
	}
	// The leading decade; the earliest wins a tie, and ties counts them.
	topDecade, topN, ties := 0, 0, 0
	for d, n := range decades {
		if n > topN || (n == topN && d < topDecade) {
			topDecade, topN = d, n
		}
	}
	for _, n := range decades {
		if n == topN {
			ties++
		}
	}
	when := &tasteLine{}
	switch {
	case topN == 5:
		when.t("All five are from the %ds. %s live there now.", topDecade, v.S)
	case topN >= 3:
		when.t("%s live in the %ds: %s of the five come from that decade.", v.S, topDecade, spell(topN))
	case topN == 2 && ties == 1:
		when.t("The %ds get two seats; no other decade gets more than one.", topDecade)
	case topN == 2:
		var ds []int
		for d, n := range decades {
			if n == 2 {
				ds = append(ds, d)
			}
		}
		sort.Ints(ds)
		when.t("Split down the middle between the %ds and the %ds, two films apiece.", ds[0], ds[1])
	case len(decades) >= 4:
		when.t("%s films, %s different decades. %s don’t do eras.", capitalize(spell(len(dated))), spell(len(decades)), v.S)
	}
	if len(dated) >= 2 {
		sort.SliceStable(dated, func(i, j int) bool { return releaseYear(dated[i]) < releaseYear(dated[j]) })
		old, nu := dated[0], dated[len(dated)-1]
		gap := releaseYear(nu) - releaseYear(old)
		if len(when.p) > 0 {
			when.t(" ")
		}
		switch {
		case gap == 0:
			when.t("Every one of them is from %d. A very good year.", releaseYear(old))
		case gap >= 40:
			when.t("It runs from ").film(old).t(" (%d) to ", releaseYear(old)).film(nu).
				t(" (%d): %d years of cinema in five films.", releaseYear(nu), gap)
		case gap <= 10:
			when.t("It all fits inside %d years, from ", gap).film(old).t(" (%d) to ", releaseYear(old)).film(nu).t(" (%d).", releaseYear(nu))
		default:
			when.t("Oldest is ").film(old).t(" (%d), newest ", releaseYear(old)).film(nu).t(" (%d).", releaseYear(nu))
		}
	}
	add(when)

	first := in.Voters <= 1
	if first {
		add((&tasteLine{}).t("This is the first ballot Filmstop has counted, so for now the whole list is %s five. Every film on it is a secret, for the moment.", v.P))
		t.Headline = "The first five in."
		return t
	}

	// Mainstream or contrarian, against the current list.
	var on, off, secrets []store.Movie
	for _, m := range in.Picks {
		if in.Places[m.ID] > 0 {
			on = append(on, m)
		} else {
			off = append(off, m)
		}
		if in.Counts[m.ID] <= 1 {
			secrets = append(secrets, m)
		}
	}
	crowd := &tasteLine{}
	switch len(on) {
	case 5:
		crowd.t("All five are on Filmstop’s list. %s the consensus, and the consensus is lucky to have %s.", capitalize(v.Re), v.O)
	case 4:
		crowd.t("Four of the five are on Filmstop’s list. The odd one out is ").film(off[0]).t(".")
	case 3:
		crowd.t("Three of the five are on Filmstop’s list; ").films(off, "and").t(" are %s own business.", v.P)
	case 2, 1:
		crowd.t("Only %s of the five %s made Filmstop’s list. The rest is %s own business.",
			spell(len(on)), map[bool]string{true: "has", false: "have"}[len(on) == 1], v.P)
	case 0:
		crowd.t("Not one of these is on Filmstop’s list. Gloriously out of step.")
	}
	one := in.Picks[0]
	switch p := in.Places[one.ID]; {
	case p == 1:
		crowd.t(" And ").film(one).t(" is Filmstop’s number one too. Great minds.")
	case p > 0:
		crowd.t(" %s number one, ", capitalize(v.P)).film(one).t(", sits at No. %d on the list.", p)
	case len(on) > 0:
		crowd.t(" Filmstop hasn’t caught up with %s number one yet.", v.P)
	}
	if in.BoardOne.ID != 0 && in.Places[one.ID] != 1 && !slices.ContainsFunc(in.Picks, func(m store.Movie) bool { return m.ID == in.BoardOne.ID }) {
		crowd.t(" Filmstop’s own number one, ").film(in.BoardOne).t(", didn’t make the cut.")
	}
	add(crowd)

	// Secrets: films nobody else has picked.
	secret := &tasteLine{}
	switch len(secrets) {
	case 0:
	case 1:
		secret.t("Nobody else has picked ").film(secrets[0]).t(". That one is %s secret.", v.P)
	case 5:
		secret.t("Nobody else has picked a single one of these. A private screening.")
	default:
		secret.t("Nobody else has picked ").films(secrets, "or").t(". Those are %s secrets.", v.P)
	}
	add(secret)

	// Kindred spirit: the closest other ballot, by overlap only.
	var kin []store.Movie
	for _, m := range in.Picks {
		if slices.Contains(in.Kindred, m.ID) {
			kin = append(kin, m)
		}
	}
	kp := &tasteLine{}
	switch len(kin) {
	case 0: // then all five are secrets, already said
	case 1:
		kp.t("The closest anyone comes is one shared film: ").film(kin[0]).t(".")
	case 5:
		kp.t("Somebody else picked the very same five. %s have a twin out there.", v.S)
	default:
		kp.t("Somewhere out there is a ballot that shares %s of these: ", spell(len(kin))).films(kin, "and").t(". A kindred spirit.")
	}
	add(kp)

	switch {
	case len(secrets) == 5:
		t.Headline = "A private screening."
	case len(kin) == 5:
		t.Headline = "Great minds, exactly."
	case len(on) == 5:
		t.Headline = "Straight down the middle of the canon."
	case topN >= 3:
		t.Headline = fmt.Sprintf("Firmly set in the %ds.", topDecade)
	case len(on) == 0:
		t.Headline = "Out of step, on purpose."
	case len(decades) == 5:
		t.Headline = "Five decades, no favourites."
	case len(kin) >= 3:
		t.Headline = "Not alone in this."
	case len(on) <= 2:
		t.Headline = fmt.Sprintf("A bit of the canon, a lot of %s.", v.O)
	default:
		t.Headline = fmt.Sprintf("Mostly the canon, with %s own accent.", v.P)
	}
	return t
}
