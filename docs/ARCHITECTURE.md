# Architecture

## Shape

One Go process serves everything. `main.go` reads the environment, opens the store, builds the TMDB client, and hands both to `web.New`, which returns an `http.Handler`. There is no separate frontend: templates, CSS, fonts and htmx are embedded in the binary with `go:embed`.

```
browser ── HTML / htmx fragments ──▶ Gin (internal/web)
                                       │
                 ┌─────────────────────┼──────────────────────┐
                 ▼                     ▼                      ▼
          boardKeeper            internal/store          internal/tmdb
       (the list in memory)   (SQLite file or Turso)   (search + details, cached)
                 │                     ▲
                 └── internal/rank ────┘
```

## A request

1. **Voter.** Every route except `/static/*` and `/healthz` runs the voter middleware (`web/voter.go`). It reads the `fs_voter` cookie, a random 32-hex id plus an HMAC-SHA256 signature, and issues a new one if it's missing or doesn't verify. The signing secret is created on first run and kept in the database (`meta` table), so cookies survive restarts.
2. **Origin check.** A POST whose `Origin` header names another host gets a 403 (`sameOrigin`). SameSite=Lax already keeps the cookie off cross-site posts; this covers browsers that ignore it.
3. **Handler.** Pages render through `html/template`. A ballot action (`pick`, `move`, `remove`, `cast`) changes the draft in one transaction. For htmx requests (`HX-Request` header) it then answers with the new storyboard fragment. Plain form posts are redirected back to the page instead, so everything works without JavaScript.
4. **Response.** Pages are sent `Cache-Control: no-store` with `Vary: Accept-Encoding, HX-Request`, and gzipped when the client accepts it and the body is at least 1 KB.

## The list

The ranked list lives in memory, not in the database. `boardKeeper` (`web/board.go`) holds the current `Board` behind an `atomic.Pointer`, so a page read is a pointer load.

- **Rebuild.** Casting a ballot calls `Touch`, which drops a token in a channel with room for one. A background loop takes the token, reads every film's tally from the store, orders them with `internal/rank`, and swaps in a new `Board` with the top 30, the voter count, the programme notes and the latest ballots. A burst of casts collapses into one rebuild.
- **Details.** After a rebuild, films on the board without TMDB details (director, runtime, genres, tagline, stills) are fetched four at a time and stored. A film whose fetch failed waits an hour before the next try.
- **Home page cache.** The rendered home page is cached per board and per minute (relative times like "3 min ago" change each minute), in two versions: for visitors who have and haven't voted. Each version is gzipped once.
- **Startup.** The first board is built before the server starts listening. Because the board is only rebuilt by this process, a change made straight to the database (seeding production, say) shows up only after a restart.

## Scoring

`internal/rank` is pure functions with no I/O. `Points(position)` gives 5 for first place down to 1 for fifth. `Order` sums points and breaks ties by more first places, then more ballots, then the lower TMDB id, so the order is stable. The package comment and `pl_test.go` explain why positional scoring beats Plackett–Luce and Condorcet for top-five ballots.

## Data

All tables are created on startup with `CREATE TABLE IF NOT EXISTS`, and columns added since the first release are added by `migrate` in `store/store.go`.

| Table | Holds |
| --- | --- |
| `meta` | Key–value settings: the cookie-signing secret |
| `movies` | Films seen in search or ballots: title, year, poster and backdrop paths, overview, and the TMDB details once fetched (`detailed` = 1) |
| `ballots` | Counted ballots: `(voter, position 1–5, movie_id)` |
| `ballot_times` | When each voter last cast; feeds "latest ballots" |
| `drafts` | The ballot being edited, same shape as `ballots` |
| `draft_open` | Voters who have a draft, even an empty one, so clearing every scene doesn't bring the counted ballot back |
| `share_codes` | One short code per voter for `/five/:code`: 10 characters from a-z and 2-9 without i, l and o |

Casting copies the draft into `ballots`, clears the draft and stamps `ballot_times`, all in one transaction.

### SQLite or Turso

`store.Source` picks the database: `TURSO_DATABASE_URL` if it's set, otherwise the `DATABASE_PATH` file.

- **File:** `modernc.org/sqlite` (pure Go) in WAL mode with `busy_timeout`, and transactions that take the write lock up front (`_txlock=immediate`), so concurrent writers wait instead of failing halfway.
- **Turso:** `libsql-client-go`, which sends each statement as an HTTP POST. Connections that sit idle can die silently on the way to Turso, and `net/http` never retries a POST, so `store/remote.go` installs a transport that drops idle connections after 20 seconds and resends a request once when it failed on a reused, already-dead connection.

## TMDB

`internal/tmdb` makes two calls: search (`/search/movie`) and details (`/movie/:id` with credits). Search results are cached in memory for 30 minutes. A v3 key goes in the query string and a v4 token in the `Authorization` header. Errors never include the request URL, because it can carry the key. Without a key, `Enabled` reports false and the site runs with search off.

## Static files

Files in `web/static` are loaded at startup, named by a short SHA-256 of their content, gzipped once and served with `Cache-Control: public, max-age=31536000, immutable`. Changing a file changes its URL, so browsers never see a stale version. CSS for the page being rendered is inlined, and the fonts are preloaded.
