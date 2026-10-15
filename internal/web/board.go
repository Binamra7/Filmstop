package web

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Binamra7/filmstop/internal/rank"
	"github.com/Binamra7/filmstop/internal/store"
	"github.com/Binamra7/filmstop/internal/tmdb"
)

const (
	boardSize   = 30
	arrivalsMax = 5
	// detailsRetry is how long a film whose details failed waits before enrich
	// asks TMDB again, so one bad id doesn't cost a request on every cast.
	detailsRetry = time.Hour
)

type Entry struct {
	rank.Standing
	store.Movie
}

// Board is the ranked list as of one moment. It's rebuilt off the request path
// after each cast, so reading it is a pointer load.
type Board struct {
	Entries []Entry
	Voters  int
	At      time.Time
	// Notes and Latest give the home page its voice; both are worked out here,
	// once per rebuild, never per request.
	Notes  []Note
	Latest []Arrival
}

func (b *Board) Top() []Entry {
	return b.Entries[:min(rank.Slots, len(b.Entries))]
}

func (b *Board) Rest() []Entry {
	if len(b.Entries) <= rank.Slots {
		return nil
	}
	return b.Entries[rank.Slots:]
}

type boardKeeper struct {
	st    *store.Store
	tmdb  *tmdb.Client
	cur   atomic.Pointer[Board]
	dirty chan struct{}

	mu     sync.Mutex
	failed map[int64]time.Time // film id -> when its details last failed
}

func newBoardKeeper(ctx context.Context, st *store.Store, tm *tmdb.Client) (*boardKeeper, error) {
	k := &boardKeeper{st: st, tmdb: tm, dirty: make(chan struct{}, 1), failed: map[int64]time.Time{}}
	if err := k.rebuild(ctx); err != nil {
		return nil, err
	}
	go k.loop(ctx)
	k.Touch() // fill in details for films already on the board
	return k, nil
}

func (k *boardKeeper) Load() *Board { return k.cur.Load() }

// Touch asks for a rebuild. A burst of casts collapses into one rebuild.
func (k *boardKeeper) Touch() {
	select {
	case k.dirty <- struct{}{}:
	default:
	}
}

func (k *boardKeeper) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-k.dirty:
			if err := k.rebuild(ctx); err != nil {
				slog.Error("rebuild board", "err", err)
				continue
			}
			if k.enrich(ctx) > 0 {
				if err := k.rebuild(ctx); err != nil {
					slog.Error("rebuild board", "err", err)
				}
			}
		}
	}
}

// enrich fetches TMDB details for films on the board that don't have them yet,
// four at a time, and returns how many it filled in.
func (k *boardKeeper) enrich(ctx context.Context) int {
	if !k.tmdb.Enabled() {
		return 0
	}
	var (
		wg   sync.WaitGroup
		sem  = make(chan struct{}, 4)
		done atomic.Int32
	)
	for _, e := range k.Load().Entries {
		k.mu.Lock()
		at, failed := k.failed[e.MovieID]
		k.mu.Unlock()
		if e.Detailed || failed && time.Since(at) < detailsRetry {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(id int64) {
			defer func() { <-sem; wg.Done() }()
			backdrop, d, err := k.tmdb.Details(ctx, id)
			if err == nil {
				err = k.st.SetDetails(ctx, id, backdrop, d)
			}
			k.mu.Lock()
			defer k.mu.Unlock()
			if err != nil {
				slog.Warn("film details", "id", id, "err", err)
				k.failed[id] = time.Now()
				return
			}
			delete(k.failed, id)
			done.Add(1)
		}(e.MovieID)
	}
	wg.Wait()
	return int(done.Load())
}

func (k *boardKeeper) rebuild(ctx context.Context) error {
	tallies, voters, err := k.st.Tallies(ctx)
	if err != nil {
		return err
	}
	standings := rank.Order(tallies)
	standings = standings[:min(boardSize, len(standings))]
	ids := make([]int64, len(standings))
	for i, s := range standings {
		ids[i] = s.MovieID
	}
	movies, err := k.st.Movies(ctx, ids)
	if err != nil {
		return err
	}
	openers, err := k.st.LatestOpeners(ctx, arrivalsMax)
	if err != nil {
		return err
	}
	b := &Board{Voters: voters, At: time.Now(), Entries: make([]Entry, len(standings)), Latest: arrivals(openers)}
	for i, s := range standings {
		b.Entries[i] = Entry{Standing: s, Movie: movies[s.MovieID]}
	}
	b.Notes = programmeNotes(b.Entries, voters)
	k.cur.Store(b)
	return nil
}
