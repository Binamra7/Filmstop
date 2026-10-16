// Package store keeps films, ballots, drafts and share codes in SQLite: a local
// file, or a Turso (libSQL) database when TURSO_DATABASE_URL is set.
package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Binamra7/filmstop/internal/rank"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

// Movie is a film as stored from TMDB. ID is TMDB's own id.
type Movie struct {
	ID           int64
	Title        string
	Year         string
	PosterPath   string
	BackdropPath string // a 16:9 still from the film
	Overview     string
	Details
}

// Details come from a second TMDB call per film, made when the film reaches the
// board, its page is opened or it is on a shared card.
type Details struct {
	Tagline  string
	Director string // up to two names, joined by " & "
	Runtime  int    // minutes
	Genres   string // up to two, comma-separated
	Detailed bool   // fetched from TMDB; false means fetch again
}

// Ballot is five movie ids in rank order; 0 marks an empty slot.
type Ballot [rank.Slots]int64

// Full reports whether every slot holds a film.
func (b Ballot) Full() bool {
	for _, id := range b {
		if id == 0 {
			return false
		}
	}
	return true
}

// Store is the database. It is safe for concurrent use.
type Store struct{ db *sql.DB }

const schema = `
CREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS movies (
	id INTEGER PRIMARY KEY,
	title TEXT NOT NULL,
	year TEXT NOT NULL DEFAULT '',
	poster_path TEXT NOT NULL DEFAULT '',
	overview TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS ballots (
	voter TEXT NOT NULL,
	position INTEGER NOT NULL CHECK (position BETWEEN 1 AND 5),
	movie_id INTEGER NOT NULL REFERENCES movies(id),
	PRIMARY KEY (voter, position)
) WITHOUT ROWID;
CREATE INDEX IF NOT EXISTS ballots_movie ON ballots(movie_id, position);
CREATE TABLE IF NOT EXISTS ballot_times (voter TEXT PRIMARY KEY, cast_at INTEGER NOT NULL);
CREATE INDEX IF NOT EXISTS ballot_times_cast_at ON ballot_times(cast_at);
CREATE TABLE IF NOT EXISTS drafts (
	voter TEXT NOT NULL,
	position INTEGER NOT NULL CHECK (position BETWEEN 1 AND 5),
	movie_id INTEGER NOT NULL REFERENCES movies(id),
	PRIMARY KEY (voter, position)
) WITHOUT ROWID;
-- A voter with a row here has a draft, even an empty one, so clearing every
-- scene doesn't bring the cast ballot back.
CREATE TABLE IF NOT EXISTS draft_open (voter TEXT PRIMARY KEY) WITHOUT ROWID;
`

// columns added after the first release; Open adds any that are missing.
var migrations = []struct{ name, def string }{
	{"backdrop_path", "TEXT NOT NULL DEFAULT ''"},
	{"tagline", "TEXT NOT NULL DEFAULT ''"},
	{"director", "TEXT NOT NULL DEFAULT ''"},
	{"runtime", "INTEGER NOT NULL DEFAULT 0"},
	{"genres", "TEXT NOT NULL DEFAULT ''"},
	{"detailed", "INTEGER NOT NULL DEFAULT 0"},
}

func migrate(db *sql.DB) error {
	rows, err := db.Query(`SELECT name FROM pragma_table_info('movies')`)
	if err != nil {
		return err
	}
	have := map[string]bool{}
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			rows.Close()
			return err
		}
		have[n] = true
	}
	rows.Close()
	for _, m := range migrations {
		if !have[m.name] {
			if _, err := db.Exec(`ALTER TABLE movies ADD COLUMN ` + m.name + ` ` + m.def); err != nil {
				return err
			}
		}
	}
	return nil
}

// Source picks the database: Turso when TURSO_DATABASE_URL is set (hosts whose
// disk is wiped on restart), otherwise the SQLite file at path.
func Source(path string) string {
	u := os.Getenv("TURSO_DATABASE_URL")
	if u == "" {
		return path
	}
	if t := os.Getenv("TURSO_AUTH_TOKEN"); t != "" {
		u += "?authToken=" + url.QueryEscape(t)
	}
	return u
}

func remote(src string) bool {
	for _, p := range []string{"libsql://", "https://", "http://"} {
		if strings.HasPrefix(src, p) {
			return true
		}
	}
	return false
}

// Open connects to src, a SQLite file path or a Turso URL from Source, and
// creates or updates the schema.
func Open(src string) (*Store, error) {
	driver, dsn := "libsql", src
	if !remote(src) {
		// WAL lets readers run in parallel. Transactions take the write lock up
		// front (_txlock=immediate), so concurrent writers queue on busy_timeout
		// instead of failing with SQLITE_BUSY halfway through.
		driver = "sqlite"
		dsn = "file:" + src + "?_txlock=immediate&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)"
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	if _, err := db.Exec(schema + fiveSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("schema: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// Secret returns the cookie-signing key, creating it on first run so voters keep
// their identity across restarts.
func (s *Store) Secret(ctx context.Context) ([]byte, error) {
	var v string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM meta WHERE key = 'secret'`).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		buf := make([]byte, 32)
		rand.Read(buf)
		v = hex.EncodeToString(buf)
		_, err = s.db.ExecContext(ctx, `INSERT INTO meta (key, value) VALUES ('secret', ?)`, v)
	}
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(v)
}

// UpsertMovies adds films or updates their title, year and overview. An empty
// poster or backdrop never replaces a stored one, and details are left alone.
func (s *Store) UpsertMovies(ctx context.Context, ms []Movie) error {
	if len(ms) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO movies (id, title, year, poster_path, backdrop_path, overview)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET title = excluded.title, year = excluded.year,
			poster_path = CASE WHEN excluded.poster_path = '' THEN poster_path ELSE excluded.poster_path END,
			overview = excluded.overview,
			backdrop_path = CASE WHEN excluded.backdrop_path = '' THEN backdrop_path ELSE excluded.backdrop_path END`)
	if err != nil {
		return err
	}
	for _, m := range ms {
		if _, err := stmt.ExecContext(ctx, m.ID, m.Title, m.Year, m.PosterPath, m.BackdropPath, m.Overview); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Movies returns the films for the given ids, keyed by id.
func (s *Store) Movies(ctx context.Context, ids []int64) (map[int64]Movie, error) {
	out := make(map[int64]Movie, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	q := `SELECT id, title, year, poster_path, backdrop_path, overview,
		tagline, director, runtime, genres, detailed FROM movies WHERE id IN (?` +
		strings.Repeat(",?", len(ids)-1) + `)`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.PosterPath, &m.BackdropPath, &m.Overview,
			&m.Tagline, &m.Director, &m.Runtime, &m.Genres, &m.Detailed); err != nil {
			return nil, err
		}
		out[m.ID] = m
	}
	return out, rows.Err()
}

// SetDetails stores a film's details, and its backdrop and poster if TMDB gave
// them. Only details with Detailed set (those from TMDB) stop the server
// fetching them.
func (s *Store) SetDetails(ctx context.Context, id int64, backdrop, poster string, d Details) error {
	detailed := 0
	if d.Detailed {
		detailed = 1
	}
	_, err := s.db.ExecContext(ctx, `UPDATE movies SET tagline = ?, director = ?, runtime = ?, genres = ?, detailed = ?,
		backdrop_path = CASE WHEN ? = '' THEN backdrop_path ELSE ? END,
		poster_path = CASE WHEN ? = '' THEN poster_path ELSE ? END WHERE id = ?`,
		d.Tagline, d.Director, d.Runtime, d.Genres, detailed, backdrop, backdrop, poster, poster, id)
	return err
}

// querier is a *sql.DB or *sql.Tx.
type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func readBallot(ctx context.Context, q querier, table, voter string) (Ballot, error) {
	var b Ballot
	rows, err := q.QueryContext(ctx, `SELECT position, movie_id FROM `+table+` WHERE voter = ?`, voter)
	if err != nil {
		return b, err
	}
	defer rows.Close()
	for rows.Next() {
		var pos int
		var id int64
		if err := rows.Scan(&pos, &id); err != nil {
			return b, err
		}
		b[pos-1] = id
	}
	return b, rows.Err()
}

func writeBallot(ctx context.Context, tx *sql.Tx, table, voter string, b Ballot) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE voter = ?`, voter); err != nil {
		return err
	}
	for i, id := range b {
		if id == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO `+table+` (voter, position, movie_id) VALUES (?, ?, ?)`, voter, i+1, id); err != nil {
			return err
		}
	}
	return nil
}

// Draft returns the voter's work-in-progress ballot. With no draft it starts
// from their cast ballot, so changing your mind begins from what you voted.
func (s *Store) Draft(ctx context.Context, voter string) (Ballot, error) {
	return draft(ctx, s.db, voter)
}

func draft(ctx context.Context, q querier, voter string) (Ballot, error) {
	b, err := readBallot(ctx, q, "drafts", voter)
	if err != nil || b != (Ballot{}) {
		return b, err
	}
	var open bool
	if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM draft_open WHERE voter = ?)`, voter).Scan(&open); err != nil || open {
		return b, err
	}
	return readBallot(ctx, q, "ballots", voter)
}

func saveDraft(ctx context.Context, tx *sql.Tx, voter string, b Ballot) error {
	if err := writeBallot(ctx, tx, "drafts", voter, b); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO draft_open (voter) VALUES (?) ON CONFLICT DO NOTHING`, voter)
	return err
}

// SaveDraft replaces the voter's draft with b. The draft stays open even when b
// is empty, so Draft won't fall back to the cast ballot.
func (s *Store) SaveDraft(ctx context.Context, voter string, b Ballot) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := saveDraft(ctx, tx, voter, b); err != nil {
		return err
	}
	return tx.Commit()
}

// EditDraft applies fn to the voter's draft and saves the result, all in one
// transaction, so two edits at once can't overwrite each other.
func (s *Store) EditDraft(ctx context.Context, voter string, fn func(b *Ballot)) (Ballot, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Ballot{}, err
	}
	defer tx.Rollback()
	b, err := draft(ctx, tx, voter)
	if err != nil {
		return b, err
	}
	before := b
	fn(&b)
	if b == before {
		return b, nil
	}
	if err := saveDraft(ctx, tx, voter, b); err != nil {
		return b, err
	}
	return b, tx.Commit()
}

// ClearVoter deletes the voter's counted ballot and draft.
func (s *Store) ClearVoter(ctx context.Context, voter string) error {
	return s.clearVoters(ctx, `voter = ?`, voter)
}

// ClearVotersWithPrefix deletes the counted ballots and drafts of every voter
// whose id starts with prefix (the seed's made-up voters, say).
func (s *Store) ClearVotersWithPrefix(ctx context.Context, prefix string) error {
	return s.clearVoters(ctx, `substr(voter, 1, ?) = ?`, len(prefix), prefix)
}

func (s *Store) clearVoters(ctx context.Context, where string, args ...any) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, t := range []string{"ballots", "ballot_times", "drafts", "draft_open"} {
		if _, err := tx.ExecContext(ctx, `DELETE FROM `+t+` WHERE `+where, args...); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Cast replaces the voter's counted ballot with b and clears their draft.
func (s *Store) Cast(ctx context.Context, voter string, b Ballot) error {
	return s.CastAt(ctx, voter, b, time.Now())
}

// CastAt is Cast with the ballot stamped at the given time, so seeded ballots
// can arrive over days rather than all in one second.
func (s *Store) CastAt(ctx context.Context, voter string, b Ballot, at time.Time) error {
	if !b.Full() {
		return errors.New("ballot needs five films")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := writeBallot(ctx, tx, "ballots", voter, b); err != nil {
		return err
	}
	for _, q := range []string{`DELETE FROM drafts WHERE voter = ?`, `DELETE FROM draft_open WHERE voter = ?`,
		// Delete and insert rather than upsert, so a recast gets a new rowid:
		// LatestOpeners breaks same-second ties on it.
		`DELETE FROM ballot_times WHERE voter = ?`} {
		if _, err := tx.ExecContext(ctx, q, voter); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO ballot_times (voter, cast_at) VALUES (?, ?)`, voter, at.Unix()); err != nil {
		return err
	}
	return tx.Commit()
}

// HasCast reports whether the voter has a counted ballot.
func (s *Store) HasCast(ctx context.Context, voter string) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM ballot_times WHERE voter = ?`, voter).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// Voters counts the cast ballots.
func (s *Store) Voters(ctx context.Context) (n int, err error) {
	err = s.db.QueryRowContext(ctx, `SELECT count(*) FROM ballot_times`).Scan(&n)
	return n, err
}

// Tallies counts, per film, how many ballots placed it at each position, and
// returns the number of ballots cast.
func (s *Store) Tallies(ctx context.Context) (ts []rank.Tally, voters int, err error) {
	if voters, err = s.Voters(ctx); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT movie_id,
		sum(position = 1), sum(position = 2), sum(position = 3), sum(position = 4), sum(position = 5)
		FROM ballots GROUP BY movie_id`)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var t rank.Tally
		if err := rows.Scan(&t.MovieID, &t.ByPos[0], &t.ByPos[1], &t.ByPos[2], &t.ByPos[3], &t.ByPos[4]); err != nil {
			return nil, 0, err
		}
		ts = append(ts, t)
	}
	return ts, voters, rows.Err()
}
