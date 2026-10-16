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
- **Details.** After a rebuild, films on the board whose details aren't `Complete()` are fetched four at a time and stored. Details are the director, runtime, genres, tagline, stills and the poster's colour swatch; `Details.Complete()` (`store/store.go`) is true once the details are fetched (`detailed` = 1) and the swatch is measured. A film whose fetch failed waits an hour before the next try. Film pages and cards fetch details the same way for films that aren't on the board.
- **Home page cache.** The rendered home page is cached per board and per minute (relative times like "3 min ago" change each minute), in two versions: for visitors who have and haven't voted. Each version is gzipped once.
- **Startup.** The first board is built before the server starts listening. Because the board is only rebuilt by this process, a change made straight to the database (seeding production, say) shows up only after a restart.

## Scoring

`internal/rank` is pure functions with no I/O. `Points(position)` gives 5 for first place down to 1 for fifth. `Order` sums points and breaks ties by more first places, then more ballots, then the lower TMDB id, so the order is stable. The package comment and `pl_test.go` explain why positional scoring beats Plackett–Luce and Condorcet for top-five ballots.

## Data

All tables are created on startup with `CREATE TABLE IF NOT EXISTS`, and columns added since the first release are added by `migrate` in `store/store.go`.

| Table | Holds |
| --- | --- |
| `meta` | Key–value settings: the cookie-signing secret |
| `movies` | Films seen in search or ballots: title, year, poster and backdrop paths, overview, the TMDB details once fetched (`detailed` = 1), and `swatch`, the poster's dominant colour (`#rrggbb`, `none` when there's no readable poster, empty until measured) |
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

The details call also measures the poster's colour (`tmdb/swatch.go`). It downloads the poster at TMDB's smallest size (`w92`, 92 px wide) and finds its dominant colour: each pixel votes for one of 36 hue bins with a weight of saturation × value, so a small saturated title beats a large grey sky, and pixels with almost no colour don't vote. The winning bin's weighted average colour is the swatch. A poster with almost no vivid pixels (black and white) gives its plain average colour. When there's no poster or it can't be read, the swatch is `"none"`, so the film isn't measured again on every rebuild.

## Colour fields

Every film is drawn on one of eight colour fields: `orange`, `mustard`, `rose`, `aqua`, `red`, `teal`, `navy` and `ink`. `web.fieldOf` (`web/field.go`, the `field` template function) maps a swatch to a field when the page renders, by hue, with grey and near-black posters going to `ink`. A film not measured yet (empty swatch, `none`, or anything unparseable) gets a field chosen from its id, so it keeps the same colour until it's measured. Templates write the result as a class (`field-orange`); the classes in `static/site.css` set `--field`, `--on`, `--on-2` and `--shape`. [`DESIGN.md`](DESIGN.md) lists the colours and their contrast.

## Static files

Files in `web/static` are loaded at startup, named by a short SHA-256 of their content, gzipped once and served with `Cache-Control: public, max-age=31536000, immutable`. Changing a file changes its URL, so browsers never see a stale version.

- **CSS:** every `static/*.css` is inlined into every page's `<style>`, `site.css` first and then the rest in name order (`loadCSS` in `web/server.go`). There are no stylesheet requests, but each page file's rules reach every page, so they are scoped under the page's wrapper class (`.page-list` on the list's `<body>`, `.booth`, `.film-page`, `.five-page`, `.method-page`, `.nf-page`) or use class names only that page has (the ballot's `bl-` prefix).
- **Fonts:** one variable family, Bricolage Grotesque, in two files declared in `templates/layout.html`: `bricolage.woff2` (Latin, preloaded) and `bricolage-ext.woff2` (Latin Extended). Their `unicode-range`s mean the second is downloaded only when a page contains one of its characters, such as a title with an accented letter outside Latin-1.

## Motion

All motion is CSS; the only script involved removes the cast sequence early.

- **View transitions.** `site.css` sets `@view-transition { navigation: auto; }`, so in browsers that support cross-document view transitions (Chromium) moving between pages cross-fades; elsewhere pages load normally. On the ballot, the htmx actions swap with `transition:true` and each strip carries `view-transition-name: film-<id>`, so a picked strip slides in, a removed one slides out and reordered strips move to their new places. Without support the fragment just swaps.
- **The cast sequence.** Casting returns the storyboard with a `.bl-titles` overlay (`templates/partials.html`): bars in each film's field sweep across, the five titles roll and "That's a wrap" lands, about 2.8 s, all CSS animation (`static/ballot.css`). A small script in `templates/ballot.html` removes it on any click or key, or when its last animation ends.
- **Reduced motion.** Under `prefers-reduced-motion: reduce`, `site.css` turns off every animation, transition and view-transition animation, and `ballot.css` hides the cast sequence entirely.
