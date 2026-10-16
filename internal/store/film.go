package store

import (
	"context"

	"github.com/Binamra7/filmstop/internal/rank"
)

// Companion is a film that shares ballots with another, and on how many.
type Companion struct {
	Movie
	Shared int
}

// FilmTally counts how many cast ballots placed one film at each position.
func (s *Store) FilmTally(ctx context.Context, id int64) (rank.Tally, error) {
	t := rank.Tally{MovieID: id}
	err := s.db.QueryRowContext(ctx, `SELECT
		coalesce(sum(position = 1), 0), coalesce(sum(position = 2), 0), coalesce(sum(position = 3), 0),
		coalesce(sum(position = 4), 0), coalesce(sum(position = 5), 0)
		FROM ballots WHERE movie_id = ?`, id).Scan(&t.ByPos[0], &t.ByPos[1], &t.ByPos[2], &t.ByPos[3], &t.ByPos[4])
	return t, err
}

// Companions returns the films most often cast on the same ballots as id. Ties
// go to the film its voters ranked higher, then the lower id.
func (s *Store) Companions(ctx context.Context, id int64, limit int) ([]Companion, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT m.id, m.title, m.year, m.poster_path, m.backdrop_path, m.overview,
		m.tagline, m.director, m.runtime, m.genres, m.detailed, count(*) AS shared
		FROM ballots a
		JOIN ballots b ON b.voter = a.voter AND b.movie_id != a.movie_id
		JOIN movies m ON m.id = b.movie_id
		WHERE a.movie_id = ?
		GROUP BY m.id
		ORDER BY shared DESC, sum(6 - b.position) DESC, m.id
		LIMIT ?`, id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Companion
	for rows.Next() {
		var c Companion
		if err := rows.Scan(&c.ID, &c.Title, &c.Year, &c.PosterPath, &c.BackdropPath, &c.Overview,
			&c.Tagline, &c.Director, &c.Runtime, &c.Genres, &c.Detailed, &c.Shared); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
