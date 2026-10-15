package store

import (
	"context"
	"strings"
)

// OthersWith counts, for each film, the counted ballots other than voter's that
// include it, in one query. Films nobody else picked are absent from the map.
func (s *Store) OthersWith(ctx context.Context, voter string, ids []int64) (map[int64]int, error) {
	out := make(map[int64]int, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	args := make([]any, 0, len(ids)+1)
	for _, id := range ids {
		args = append(args, id)
	}
	args = append(args, voter)
	rows, err := s.db.QueryContext(ctx, `SELECT movie_id, count(*) FROM ballots
		WHERE movie_id IN (?`+strings.Repeat(",?", len(ids)-1)+`) AND voter != ?
		GROUP BY movie_id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, rows.Err()
}
