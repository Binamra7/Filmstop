package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Binamra7/filmstop/internal/rank"
	"github.com/Binamra7/filmstop/internal/store"
	"github.com/gin-gonic/gin"
)

// filmRoutes registers the film page routes.
func (s *Server) filmRoutes(r gin.IRoutes) {
	r.GET("/film/:id", s.film)
}

const companions = 6

type filmView struct {
	page
	Film     store.Movie
	Standing rank.Standing // Place is 0 when the film isn't on the list
	Voters   int
	Note     string // one line on how voters place it
	Short    string // how far a film off the list is from getting on
	Also     []store.Companion
	Draft    int  // its position in the visitor's five, or 0
	Full     bool // the visitor's five has no room left
}

func (s *Server) film(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		s.filmNotFound(c)
		return
	}
	found, err := s.st.Movies(c.Request.Context(), []int64{id})
	if err != nil {
		s.fail(c, err)
		return
	}
	m, ok := found[id]
	if !ok {
		s.filmNotFound(c)
		return
	}
	if !m.Detailed && s.tmdb.Enabled() {
		s.fillDetails(c, &m)
	}

	v := filmView{page: page{m.Title + " · Filmstop", ""}, Film: m}
	b := s.board.Load()
	for _, e := range b.Entries {
		if e.MovieID == id {
			v.Standing = e.Standing
			break
		}
	}
	if v.Standing.Place == 0 {
		t, err := s.st.FilmTally(c.Request.Context(), id)
		if err != nil {
			s.fail(c, err)
			return
		}
		// Order skips films with no ballots, so an unpicked film scores nothing.
		if o := rank.Order([]rank.Tally{t}); len(o) == 1 {
			v.Standing = o[0]
			v.Standing.Place = 0
			if len(b.Entries) > 0 {
				v.Short = shortOfList(v.Standing, b.Entries[len(b.Entries)-1])
			}
		}
	}
	v.Voters = b.Voters
	v.Note = placeNote(v.Standing)
	if v.Standing.Ballots > 0 {
		if v.Also, err = s.st.Companions(c.Request.Context(), id, companions); err != nil {
			s.fail(c, err)
			return
		}
	}
	draft, err := s.st.Draft(c.Request.Context(), voterID(c))
	if err != nil {
		s.fail(c, err)
		return
	}
	v.Full = draft.Full()
	for i, have := range draft {
		if have == id {
			v.Draft = i + 1
		}
	}
	s.render(c, http.StatusOK, "film", v)
}

func (s *Server) filmNotFound(c *gin.Context) {
	s.render(c, http.StatusNotFound, "notfound", page{"Not found · Filmstop", ""})
}

// fillDetails fetches a film's director, tagline and runtime on first view. A
// slow or failed answer just leaves the page without them.
func (s *Server) fillDetails(c *gin.Context, m *store.Movie) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2500*time.Millisecond)
	defer cancel()
	backdrop, d, err := s.tmdb.Details(ctx, m.ID)
	if err == nil {
		err = s.st.SetDetails(ctx, m.ID, backdrop, d)
	}
	if err != nil {
		slog.Warn("film details", "id", m.ID, "err", err)
		return
	}
	m.Details = d
	if backdrop != "" {
		m.BackdropPath = backdrop
	}
}

// placeNote says, in a sentence, where the people who chose a film tend to put it.
func placeNote(st rank.Standing) string {
	n := st.Ballots
	switch {
	case n == 0:
		return ""
	case n == 1:
		for i, k := range st.ByPos {
			if k > 0 {
				if i == 0 {
					return "One person has it in their five, and they put it first."
				}
				return fmt.Sprintf("One person has it in their five, at number %d.", i+1)
			}
		}
	case st.Firsts == n:
		return "Everyone who chose it put it first. Nobody picks this one as a supporting player."
	case st.Firsts*2 > n:
		return "Most of the people who chose it put it first. This is a film people lead with."
	}
	mode, top, lo, hi := 0, 0, 0, 0
	for i, k := range st.ByPos {
		if k == 0 {
			continue
		}
		if lo == 0 {
			lo = i + 1
		}
		hi = i + 1
		if k > top {
			mode, top = i, k
		}
	}
	for i, k := range st.ByPos {
		if k == top && i != mode {
			if top == 1 {
				return fmt.Sprintf("No two of its voters put it in the same place: it turns up anywhere from number %d to number %d.", lo, hi)
			}
			return fmt.Sprintf("Its voters can’t agree where it belongs, anywhere from number %d to number %d.", lo, hi)
		}
	}
	switch mode {
	case 0:
		return "Number one is where it lands most often."
	case 1, 2:
		return fmt.Sprintf("It lands at number %d most often: close to the heart, a step behind the film people lead with.", mode+1)
	}
	return fmt.Sprintf("It lands at number %d most often: the film people can’t leave off, even when it isn’t the one they open with.", mode+1)
}

// shortOfList says how far a ranked-but-unlisted film is from the last place on the list.
func shortOfList(st rank.Standing, last Entry) string {
	gap := last.Points - st.Points
	if gap <= 0 {
		return fmt.Sprintf("It’s level with %s at the foot of the list; one more vote tips it in.", last.Title)
	}
	unit := "points"
	if gap == 1 {
		unit = "point"
	}
	how := "One first place would put it on."
	if gap >= rank.Points(1) {
		how = fmt.Sprintf("That’s %d first places away.", (gap+rank.Points(1))/rank.Points(1))
	}
	return fmt.Sprintf("It’s %d %s behind %s, number %d. %s", gap, unit, last.Title, last.Place, how)
}
