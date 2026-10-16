// Command seed fills a local database with made-up ballots over a built-in
// catalogue of real films, so every part of the site has something to show
// during development. Never run it against production: the ballots are fake.
// It refuses a Turso database (TURSO_DATABASE_URL) unless -remote is passed.
//
//	go run ./cmd/seed -db filmstop.db -voters 400 -days 45
//
// Rerunning it replaces the previous run's crowd and resets every persona.
//
// It needs no network. With TMDB_API_KEY set it also fetches each film's real
// details and still, so the pages have pictures.
//
// Besides the crowd it casts a few named personas and prints a cookie for each,
// so you can browse as a voter with a card, a half-finished draft, and so on.
package main

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Binamra7/filmstop/internal/store"
	"github.com/Binamra7/filmstop/internal/tmdb"
	"github.com/Binamra7/filmstop/internal/web"
)

func main() {
	dbPath := flag.String("db", cmp.Or(os.Getenv("DATABASE_PATH"), "filmstop.db"), "SQLite file (ignored when TURSO_DATABASE_URL is set)")
	remote := flag.Bool("remote", false, "allow seeding the Turso database in TURSO_DATABASE_URL (fake ballots!)")
	voters := flag.Int("voters", 400, "simulated voters")
	days := flag.Int("days", 45, "spread the ballots over this many days")
	flag.Parse()
	if *voters < 1 || *days < 1 {
		log.Fatal("-voters and -days must be positive")
	}
	// The ballots are fake, and production runs on Turso: don't write to it just
	// because TURSO_DATABASE_URL happens to be set in this shell (or -db is a URL).
	src := store.Source(*dbPath)
	if strings.Contains(src, "://") && !*remote {
		log.Fatal("refusing to seed a remote database with fake ballots (pass -remote if you mean it)")
	}
	ctx := context.Background()

	st, err := store.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	ids := make([]int64, len(catalogue))
	for i, fm := range catalogue {
		ids[i] = fm.ID
	}
	have, err := st.Movies(ctx, ids)
	if err != nil {
		log.Fatal(err)
	}
	// Only add films the database lacks: rows already there may hold TMDB's own
	// title and overview, which the catalogue's one-liners shouldn't replace.
	var missing []store.Movie
	for _, fm := range catalogue {
		if _, ok := have[fm.ID]; !ok {
			missing = append(missing, fm.Movie)
		}
	}
	if err := st.UpsertMovies(ctx, missing); err != nil {
		log.Fatal(err)
	}
	// Built-in details are stored as not yet detailed, so the server still
	// fetches the real ones (and the still) once it has a TMDB key. Films already
	// detailed from TMDB, by an earlier seed or the server, are left alone.
	td := tmdb.New(strings.TrimSpace(os.Getenv("TMDB_API_KEY")))
	var (
		fetched atomic.Int32
		wg      sync.WaitGroup
		sem     = make(chan struct{}, 4) // as many at once as the server's board
	)
	for _, fm := range catalogue {
		if have[fm.ID].Detailed {
			fetched.Add(1)
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() { <-sem; wg.Done() }()
			backdrop, poster, d := "", "", fm.Details
			if td.Enabled() {
				if b, pp, got, err := td.Details(ctx, fm.ID); err != nil {
					log.Printf("tmdb details for %s: %v (keeping the built-in ones)", fm.Title, err)
				} else {
					backdrop, poster, d = b, pp, got
					fetched.Add(1)
				}
			}
			if err := st.SetDetails(ctx, fm.ID, backdrop, poster, d); err != nil {
				log.Fatal(err)
			}
		}()
	}
	wg.Wait()

	// Drop every earlier seed voter first, so a rerun with fewer -voters
	// doesn't leave the old run's extra (and "just now") ballots counted.
	if err := st.ClearVotersWithPrefix(ctx, seedPrefix); err != nil {
		log.Fatal(err)
	}

	r := rand.New(rand.NewPCG(7, 11))
	now := time.Now()
	for v := range *voters {
		at := castTime(r, v, *voters, *days, now)
		if err := st.CastAt(ctx, fmt.Sprintf("%s%028d", seedPrefix, v), ballotFor(r, pickTribe(r)), at); err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("seeded %d ballots over %d films (%d with TMDB details)", *voters, len(catalogue), fetched.Load())

	secret, err := st.Secret(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// fs_voter is HttpOnly, so document.cookie can't replace it; edit it in DevTools.
	fmt.Println("\nPersonas. To browse as one, open DevTools > Application (Storage in Firefox) > Cookies")
	fmt.Println("on the site, set fs_voter to the persona's cookie, then reload.")
	fmt.Println()
	for _, p := range personas {
		if err := p.apply(ctx, st, now); err != nil {
			log.Fatal(err)
		}
		id := personaID(p.name)
		line := fmt.Sprintf("  %-10s %s\n             cookie %s", p.name, p.about, web.VoterCookie(secret, id))
		if p.card {
			code, err := st.ShareCode(ctx, id)
			if err != nil {
				log.Fatal(err)
			}
			line += "\n             card   /five/" + code
		}
		fmt.Println(line)
	}
}

// seedPrefix starts every made-up crowd voter's id. It isn't hex, so no real
// voter (or persona) id starts with it.
const seedPrefix = "seed"

// pickTribe draws a tribe weighted by share. The fallback covers shares whose
// float sum lands just under 1.
func pickTribe(r *rand.Rand) tribe {
	u, sum := r.Float64(), 0.0
	for _, t := range tribes {
		if sum += t.share; u < sum {
			return t
		}
	}
	return tribes[0]
}

// ballotFor draws five films without replacement, each with probability
// proportional to its appeal to the tribe times personal taste noise (a
// Plackett–Luce sampler via the Gumbel trick). The cult puts Withnail first.
func ballotFor(r *rand.Rand, t tribe) store.Ballot {
	type keyed struct {
		id int64
		k  float64
	}
	ks := make([]keyed, 0, len(catalogue))
	for _, fm := range catalogue {
		a := fm.appeal
		for _, tag := range fm.tags {
			if b, ok := t.boost[tag]; ok {
				a *= b
			}
		}
		k := math.Log(a) - math.Log(-math.Log(r.Float64()))
		if t.name == "cult" && fm.ID == withnail {
			k = math.Inf(1)
		}
		ks = append(ks, keyed{fm.ID, k})
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i].k > ks[j].k })
	var b store.Ballot
	for i := range b {
		b[i] = ks[i].id
	}
	return b
}

// withnail is the TMDB id of Withnail & I, the cult tribe's number one.
const withnail = 13446

// castTime spreads ballots over the last days, busier lately, and puts the
// last few minutes apart so the latest ballots read as just now, minutes ago,
// within the hour.
func castTime(r *rand.Rand, v, n, days int, now time.Time) time.Time {
	if left := n - 1 - v; left < len(recent) {
		return now.Add(-recent[left])
	}
	span := time.Duration(days) * 24 * time.Hour
	return now.Add(-time.Duration(math.Pow(r.Float64(), 1.8)*float64(span)) - 2*time.Hour)
}

// recent holds how long ago the last few ballots were cast, newest first.
var recent = []time.Duration{20 * time.Second, 3 * time.Minute, 11 * time.Minute, 38 * time.Minute, 70 * time.Minute}

// A persona is a voter you can browse as, set up to show one state of the site.
type persona struct {
	name, about string
	cast        store.Ballot // counted ballot; zero for none
	castAgo     time.Duration
	draft       store.Ballot // open draft; zero for none
	card        bool         // print a share link to their card
}

var personas = []persona{
	{name: "cinephile", about: "cast an arthouse five; has a card with a kindred ballot",
		cast: store.Ballot{843, 62, 18148, 1398, 426}, castAgo: 3 * 24 * time.Hour, card: true},
	{name: "kindred", about: "shares four films with cinephile",
		cast: store.Ballot{62, 843, 426, 18148, 346}, castAgo: 9 * 24 * time.Hour, card: true},
	{name: "decades", about: "five films from five decades",
		cast: store.Ballot{289, 426, 28, 680, 496243}, castAgo: 20 * 24 * time.Hour, card: true},
	{name: "crowd", about: "cast the crowd favourites; has an open draft swapping in Interstellar",
		cast: store.Ballot{278, 155, 238, 27205, 122}, castAgo: 6 * 24 * time.Hour, card: true,
		draft: store.Ballot{278, 155, 238, 157336, 122}},
	{name: "newcomer", about: "never cast; a draft with two of five scenes filled",
		draft: store.Ballot{129, 0, 346648, 0, 0}},
}

func (p persona) apply(ctx context.Context, st *store.Store, now time.Time) error {
	id := personaID(p.name)
	// Start from nothing, so a persona changed while browsing is reset.
	if err := st.ClearVoter(ctx, id); err != nil {
		return err
	}
	if p.cast.Full() {
		if err := st.CastAt(ctx, id, p.cast, now.Add(-p.castAgo)); err != nil {
			return err
		}
	}
	if p.draft != (store.Ballot{}) {
		return st.SaveDraft(ctx, id, p.draft)
	}
	return nil
}

// personaID is a stable 32-hex voter id, so reseeding keeps the same cookies
// working (as long as the database, and so its secret, survives).
func personaID(name string) string {
	h := sha256.Sum256([]byte("filmstop persona " + name))
	return hex.EncodeToString(h[:16])
}
