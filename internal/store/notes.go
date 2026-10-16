package store

import (
	"context"
	"time"
)

// Opener is the film someone put first on a recently cast ballot. It never
// carries the voter: the home page only says "someone".
type Opener struct {
	MovieID int64
	Title   string
	At      time.Time
}

// LatestOpeners returns the number one film from the n most recently cast
// ballots, newest first.
func (s *Store) LatestOpeners(ctx context.Context, n int) ([]Opener, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT m.id, m.title, t.cast_at
		FROM ballot_times t
		JOIN ballots b ON b.voter = t.voter AND b.position = 1
		JOIN movies m ON m.id = b.movie_id
		ORDER BY t.cast_at DESC, t.rowid DESC LIMIT ?`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Opener
	for rows.Next() {
		var o Opener
		var at int64
		if err := rows.Scan(&o.MovieID, &o.Title, &at); err != nil {
			return nil, err
		}
		o.At = time.Unix(at, 0)
		out = append(out, o)
	}
	return out, rows.Err()
}
