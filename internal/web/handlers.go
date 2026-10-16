package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Binamra7/filmstop/internal/rank"
	"github.com/Binamra7/filmstop/internal/store"
	"github.com/Binamra7/filmstop/internal/tmdb"
	"github.com/gin-gonic/gin"
)

type page struct {
	Title string
	Nav   string
}

func (s *Server) home(c *gin.Context) {
	cast, err := s.st.HasCast(c.Request.Context(), voterID(c))
	if err != nil {
		s.fail(c, err)
		return
	}
	b := s.board.Load()
	r, err := s.homePage.get(b, cast, func() (*rendered, error) {
		return s.execute("home", "layout", struct {
			page
			Board *Board
			Cast  bool
		}{page{"Filmstop: the five films everyone loves most", "list"}, b, cast})
	})
	if err != nil {
		s.fail(c, err)
		return
	}
	s.send(c, http.StatusOK, r)
}

func (s *Server) method(c *gin.Context) {
	s.render(c, http.StatusOK, "method", struct {
		page
		Voters int
	}{page{"How the list is counted · Filmstop", "method"}, s.board.Load().Voters})
}

// --- ballot

type slot struct {
	Pos   int
	Movie *store.Movie
	Note  string // for a filled scene: who else has this film in their five
	Hint  string // for an empty scene: what kind of film belongs there
}

// sceneHints describe each empty scene, best first.
var sceneHints = [rank.Slots]string{
	"The one you’d keep if all the others vanished",
	"The one that nearly took first place",
	"The one you bring up at every dinner",
	"The one you’ve seen more times than you admit",
	"The one that gets the last word",
}

var countWords = [...]string{"No", "One", "Two", "Three", "Four", "Five"}

type ballotView struct {
	Slots     []slot
	Full      bool
	Filled    int
	Cast      bool   // the voter has a counted ballot
	Flash     string // one line of feedback after an action
	Done      bool   // just cast
	CastLabel string
}

// companyNote says, warmly, how many other people share a pick. place is the
// film's place on the list, 0 if it isn't on the board.
func companyNote(others, place int) string {
	if others == 0 {
		return "No one else has picked this yet. A discovery."
	}
	with := "one other person"
	if others > 1 {
		with = fmt.Sprint(others) + " others"
	}
	switch {
	case place == 1:
		return "Filmstop’s number one too, with you and " + with + "."
	case place > 1 && place <= rank.Slots:
		return "Number " + fmt.Sprint(place) + " on Filmstop, with you and " + with + "."
	}
	return "You and " + with + " have this in their five."
}

func (s *Server) ballotView(c *gin.Context, b store.Ballot, flash string) (ballotView, error) {
	ids := make([]int64, 0, rank.Slots)
	for _, id := range b {
		if id != 0 {
			ids = append(ids, id)
		}
	}
	voter := voterID(c)
	movies, err := s.st.Movies(c.Request.Context(), ids)
	if err != nil {
		return ballotView{}, err
	}
	others, err := s.st.OthersWith(c.Request.Context(), voter, ids)
	if err != nil {
		return ballotView{}, err
	}
	place := map[int64]int{}
	for _, e := range s.board.Load().Entries {
		place[e.MovieID] = e.Place
	}
	v := ballotView{Full: b.Full(), Flash: flash}
	for i, id := range b {
		sl := slot{Pos: i + 1, Hint: sceneHints[i]}
		if m, ok := movies[id]; ok {
			sl.Movie = &m
			sl.Note = companyNote(others[id], place[id])
			v.Filled++
		}
		v.Slots = append(v.Slots, sl)
	}
	if v.Cast, err = s.st.HasCast(c.Request.Context(), voter); err != nil {
		return v, err
	}
	v.setCastLabel()
	return v, nil
}

func (v *ballotView) setCastLabel() {
	switch {
	case v.Done:
		v.CastLabel = "Counted"
	case !v.Full:
		n := rank.Slots - v.Filled
		v.CastLabel = countWords[n] + " " + map[bool]string{true: "scene", false: "scenes"}[n == 1] + " to go"
	case v.Cast:
		v.CastLabel = "Count my new five"
	default:
		v.CastLabel = "That’s my five. Count it"
	}
}

func (s *Server) ballotPage(c *gin.Context) {
	b, err := s.st.Draft(c.Request.Context(), voterID(c))
	if err != nil {
		s.fail(c, err)
		return
	}
	v, err := s.ballotView(c, b, c.Query("flash"))
	if err != nil {
		s.fail(c, err)
		return
	}
	q := strings.TrimSpace(c.Query("q"))
	res := s.results(c, q, b)
	s.render(c, http.StatusOK, "ballot", struct {
		page
		Ballot  ballotView
		Results resultsView
		Search  bool
	}{page{"Your five · Filmstop", "ballot"}, v, res, s.tmdb.Enabled()})
}

// prompts nudge someone staring at an empty search box. One shows at a time,
// changing every ten minutes.
var prompts = []string{
	"The film you’ve seen more than five times.",
	"The one that made you cry on a plane.",
	"A film you’d defend to anyone.",
	"The one you quote without meaning to.",
	"The first film that made you stay for the credits.",
	"The one you press on every new friend.",
	"A film you saw at exactly the right age.",
	"The one you can’t watch without the soundtrack following you home.",
	"The film you’d choose for a rainy Sunday, every time.",
	"The one that changed what you thought films could do.",
}

const promptEvery = 10 * time.Minute

func promptAt(t time.Time) string {
	return prompts[int(t.Unix()/int64(promptEvery.Seconds()))%len(prompts)]
}

type resultsView struct {
	Query    string
	Searched bool // the query was long enough to send
	Movies   []store.Movie
	Ideas    []store.Movie // with an empty query: Filmstop's top films, one tap to add
	Prompt   string        // with an empty query: something to think about
	Picked   map[int64]int // film id -> its position on the voter's ballot
	Err      string
	OOB      bool // rendered as an htmx out-of-band swap
}

// hitList is one list of addable films, marked with where the voter has them.
type hitList struct {
	Movies []store.Movie
	Picked map[int64]int
}

// Hits pairs ms with the voter's picks for the "hits" template, since a
// template can't build a struct itself.
func (v resultsView) Hits(ms []store.Movie) hitList { return hitList{ms, v.Picked} }

func (s *Server) results(c *gin.Context, q string, b store.Ballot) resultsView {
	v := resultsView{Query: q, Picked: map[int64]int{}}
	for i, id := range b {
		if id != 0 {
			v.Picked[id] = i + 1
		}
	}
	if q == "" {
		v.Prompt = promptAt(time.Now())
		for _, e := range s.board.Load().Top() {
			v.Ideas = append(v.Ideas, e.Movie)
		}
		return v
	}
	if len([]rune(q)) < 2 {
		return v
	}
	v.Searched = true
	ms, fresh, err := s.tmdb.Search(c.Request.Context(), q)
	switch {
	case errors.Is(err, tmdb.ErrNoKey):
		v.Err = "Search is switched off on this server: it needs a TMDB API key."
	case err != nil:
		slog.Warn("tmdb search", "q", q, "err", err)
		v.Err = "The film database didn't answer. Try again in a moment."
	default:
		// Remember what we showed, so a pick can be checked against it. Cached
		// results were stored when they were fetched.
		if fresh {
			if err := s.st.UpsertMovies(c.Request.Context(), ms); err != nil {
				slog.Error("upsert movies", "err", err)
			}
		}
		v.Movies = ms
	}
	return v
}

func (s *Server) search(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if c.GetHeader("HX-Request") == "" {
		c.Redirect(http.StatusSeeOther, "/ballot?q="+url.QueryEscape(q))
		return
	}
	b, err := s.st.Draft(c.Request.Context(), voterID(c))
	if err != nil {
		s.fail(c, err)
		return
	}
	s.fragment(c, "results", s.results(c, q, b))
}

// edit loads the draft, applies fn, saves it and answers with the ballot panel
// (htmx) or a redirect back to the ballot page (plain form post).
func (s *Server) edit(c *gin.Context, fn func(b *store.Ballot) string) {
	var flash string
	b, err := s.st.EditDraft(c.Request.Context(), voterID(c), func(b *store.Ballot) { flash = fn(b) })
	if err != nil {
		s.fail(c, err)
		return
	}
	s.respondBallot(c, b, flash, false)
}

// respondBallot answers an htmx ballot action with the new panel, plus the
// search results (or, with an empty search, the quick adds) out of band so
// their "Scene N" marks stay true. A plain form post is redirected to /ballot.
func (s *Server) respondBallot(c *gin.Context, b store.Ballot, flash string, done bool) {
	if c.GetHeader("HX-Request") == "" {
		to := "/ballot"
		if flash != "" {
			to += "?flash=" + url.QueryEscape(flash)
		}
		c.Redirect(http.StatusSeeOther, to)
		return
	}
	v, err := s.ballotView(c, b, flash)
	if err != nil {
		s.fail(c, err)
		return
	}
	v.Done = done
	v.setCastLabel()
	resp := struct {
		Ballot  ballotView
		Results *resultsView
	}{Ballot: v}
	if !done {
		r := s.results(c, strings.TrimSpace(c.PostForm("q")), b)
		r.OOB = true
		resp.Results = &r
	}
	s.fragment(c, "ballot-response", resp)
}

func (s *Server) pick(c *gin.Context) {
	id, _ := strconv.ParseInt(c.PostForm("id"), 10, 64)
	found, err := s.st.Movies(c.Request.Context(), []int64{id})
	if err != nil {
		s.fail(c, err)
		return
	}
	_, ok := found[id]
	s.edit(c, func(b *store.Ballot) string {
		if !ok {
			return "That film isn't in our list any more. Search for it again."
		}
		for i, have := range b {
			if have == id {
				return "That one’s already on your storyboard, as scene " + fmt.Sprint(i+1) + "."
			}
		}
		for i := range b {
			if b[i] == 0 {
				b[i] = id
				return ""
			}
		}
		return "Your five is full. Let one go to make room."
	})
}

func formPos(c *gin.Context) int {
	p, _ := strconv.Atoi(c.PostForm("pos"))
	if p < 1 || p > rank.Slots {
		return 0
	}
	return p
}

func (s *Server) move(c *gin.Context) {
	pos := formPos(c)
	to := pos - 1
	if c.PostForm("dir") == "down" {
		to = pos + 1
	}
	s.edit(c, func(b *store.Ballot) string {
		if pos == 0 || to < 1 || to > rank.Slots {
			return ""
		}
		b[pos-1], b[to-1] = b[to-1], b[pos-1]
		return ""
	})
}

func (s *Server) remove(c *gin.Context) {
	pos := formPos(c)
	s.edit(c, func(b *store.Ballot) string {
		if pos != 0 {
			b[pos-1] = 0
		}
		return ""
	})
}

func (s *Server) cast(c *gin.Context) {
	voter := voterID(c)
	b, err := s.st.Draft(c.Request.Context(), voter)
	if err != nil {
		s.fail(c, err)
		return
	}
	if !b.Full() {
		s.respondBallot(c, b, "Five films, please. Every scene counts.", false)
		return
	}
	if err := s.st.Cast(c.Request.Context(), voter, b); err != nil {
		s.fail(c, err)
		return
	}
	s.board.Touch()
	if c.GetHeader("HX-Request") == "" {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}
	s.respondBallot(c, b, "", true)
}

func (s *Server) fail(c *gin.Context, err error) {
	slog.Error("request", "path", c.Request.URL.Path, "err", err)
	c.String(http.StatusInternalServerError, "Something broke on our side. Try again.")
}
