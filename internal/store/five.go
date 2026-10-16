package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

// Share codes are the public face of a ballot: a short random id that can go
// in a link without giving away the voter's cookie id.
const fiveSchema = `
CREATE TABLE IF NOT EXISTS share_codes (
	voter TEXT PRIMARY KEY,
	code TEXT NOT NULL UNIQUE
) WITHOUT ROWID;
`

// CodeLen is the length of a share code.
const CodeLen = 10

// codeAlphabet is a-z and 2-9 without i, l and o, which people misread.
const codeAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"

// ValidCode reports whether code could be a share code, so junk never reaches
// the database.
func ValidCode(code string) bool {
	if len(code) != CodeLen {
		return false
	}
	for _, r := range code {
		if !strings.ContainsRune(codeAlphabet, r) {
			return false
		}
	}
	return true
}

func newCode() string {
	buf := make([]byte, CodeLen)
	rand.Read(buf)
	for i, b := range buf {
		// 31 symbols, so a tiny modulo bias; irrelevant for an unguessable link.
		buf[i] = codeAlphabet[int(b)%len(codeAlphabet)]
	}
	return string(buf)
}

// ShareCode returns the voter's share code, making one the first time.
func (s *Store) ShareCode(ctx context.Context, voter string) (string, error) {
	for range 3 {
		if _, err := s.db.ExecContext(ctx, `INSERT INTO share_codes (voter, code) VALUES (?, ?)
			ON CONFLICT DO NOTHING`, voter, newCode()); err != nil {
			return "", err
		}
		var code string
		err := s.db.QueryRowContext(ctx, `SELECT code FROM share_codes WHERE voter = ?`, voter).Scan(&code)
		if err == nil {
			return code, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
		// The random code collided with another voter's; roll again.
	}
	return "", errors.New("share code: no free code")
}

// VoterByCode finds the voter behind a share code.
func (s *Store) VoterByCode(ctx context.Context, code string) (string, bool, error) {
	var voter string
	err := s.db.QueryRowContext(ctx, `SELECT voter FROM share_codes WHERE code = ?`, code).Scan(&voter)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return voter, err == nil, err
}

// CastBallot returns the voter's counted ballot (empty if they haven't cast).
func (s *Store) CastBallot(ctx context.Context, voter string) (Ballot, error) {
	return readBallot(ctx, s.db, "ballots", voter)
}

// Kindred is the other ballot that shares the most films with a voter's.
type Kindred struct {
	Shared []int64 // the films both picked
}

// Company returns, for the voter's counted ballot, how many ballots list each
// of their films (their own included), and their kindred ballot. Ties for
// kindred go to the ballot that placed the shared films closest to where the
// voter did.
func (s *Store) Company(ctx context.Context, voter string) (counts map[int64]int, kin Kindred, err error) {
	counts = map[int64]int{}
	rows, err := s.db.QueryContext(ctx, `SELECT o.movie_id, count(*) FROM ballots b
		JOIN ballots o ON o.movie_id = b.movie_id
		WHERE b.voter = ? GROUP BY o.movie_id`, voter)
	if err != nil {
		return nil, kin, err
	}
	for rows.Next() {
		var id int64
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			rows.Close()
			return nil, kin, err
		}
		counts[id] = n
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, kin, err
	}

	// One round trip: pick the kindred voter, then list what they share.
	var ns sql.NullString
	err = s.db.QueryRowContext(ctx, `SELECT group_concat(movie_id) FROM (
			SELECT b.movie_id FROM ballots b JOIN ballots o ON o.movie_id = b.movie_id
			WHERE b.voter = ?1 AND o.voter = (
				SELECT o.voter FROM ballots b JOIN ballots o ON o.movie_id = b.movie_id AND o.voter != b.voter
				WHERE b.voter = ?1 GROUP BY o.voter
				ORDER BY count(*) DESC, sum(abs(o.position - b.position)) ASC, o.voter LIMIT 1)
			ORDER BY b.position)`, voter).Scan(&ns)
	if err != nil {
		return nil, kin, err
	}
	if ns.Valid {
		for _, f := range strings.Split(ns.String, ",") {
			id, err := strconv.ParseInt(f, 10, 64)
			if err != nil {
				return nil, kin, err
			}
			kin.Shared = append(kin.Shared, id)
		}
	}
	return counts, kin, nil
}
