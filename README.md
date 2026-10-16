# Filmstop

Everyone picks their five favourite films, in order. Filmstop counts every ballot
into one list. Live at **https://filmstop.onrender.com**.

Go + Gin on the server, htmx in the browser, SQLite on disk (Turso in production).
It ships as one binary with its templates, CSS, font and htmx embedded.

## What's on it

- **The list** (`/`): the number one as an opening scene, two to five as scene
  selection, six to thirty as end credits, plus "programme notes" written from the
  votes (photo finishes, cult favourites, the most divisive film) and the latest
  ballots.
- **Film pages** (`/film/:id`): the story, where it stands, how voters place it, and
  the films that share its ballots.
- **Your five** (`/ballot`): a five-scene storyboard with search, one-tap prompts, and
  how many others share each pick.
- **Your card** (`/five`): a shareable page of your five with a reading of your taste.

## Run it

```sh
export TMDB_API_KEY=...        # TMDB v3 key or v4 read token; search is off without it
go run .                       # http://localhost:8080
```

| Variable | Default | |
| --- | --- | --- |
| `PORT` | `8080` | |
| `DATABASE_PATH` | `filmstop.db` | SQLite file. Put it on a persistent disk |
| `TMDB_API_KEY` | none | Needed for film search |
| `TURSO_DATABASE_URL` | none | Use this Turso database instead of the SQLite file |
| `TURSO_AUTH_TOKEN` | none | Token for `TURSO_DATABASE_URL` |

For a list with something on it, seed made-up ballots (local only, never
production):

```sh
go run ./cmd/seed -db filmstop.db -voters 400 -days 45
```

It needs no network: 72 real films (with their TMDB ids) and voters in taste
clusters, cast over the last few weeks, so every programme note, film page and
"latest ballots" line has something to say. With `TMDB_API_KEY` set it also
fetches each film's real details, still and poster. It then prints cookies for a few
personas (a voter with a card, a kindred ballot, an open draft, a newcomer
halfway through) so you can browse as each one. Rerunning on the same database
replaces the seeded ballots rather than adding more.

## Docker

```sh
cp .env.example .env            # add your TMDB_API_KEY
docker compose up -d --build    # http://localhost:8080
```

The database lives in the `filmstop-data` volume, so it survives rebuilds.
`docker compose down -v` deletes it. For demo data, seed before starting (the
server reads the board at startup):

```sh
docker compose run --rm seed && docker compose up -d
```

Without compose: `docker build -t filmstop . && docker run -p 8080:8080 -v filmstop-data:/data -e TMDB_API_KEY filmstop`.

## Deploy (Render + Turso, both free)

Render's free web service wipes its disk on restart, so in production the
database is [Turso](https://turso.tech) (hosted SQLite). Without
`TURSO_DATABASE_URL` the app uses the local SQLite file as before.

1. In Turso, create a database and an auth token. Note the `libsql://...` URL.
2. In Render, choose **New > Blueprint**, pick this repo, and fill in
   `TMDB_API_KEY`, `TURSO_DATABASE_URL` and `TURSO_AUTH_TOKEN`. `render.yaml`
   does the rest.

The free service sleeps after 15 idle minutes, and the next visit takes about a
minute to wake it. Set the Turso variables in Render only, not in your local
`.env`, so local runs never touch production.

The seed's ballots are made up, so it refuses to write to a remote database
unless you pass `-remote`. Over the network it writes one ballot per transaction,
so a large remote seed is slow. After seeding, restart the service in Render: the
server reads the board at startup.

## How the list is counted

Each ballot gives 5 points to its #1 down to 1 point to its #5. Points are summed,
and ties go to more first places, then more ballots. Plackett–Luce and Condorcet
methods were tested and rejected: with top-5 ballots drawn from every film ever
made, they learn almost nothing from order (see `internal/rank/pl_test.go`). The
full reasoning is on the site's Method page.

## Layout

```
main.go              server entry point
cmd/seed             demo data loader
internal/rank        scoring and ordering (pure, tested)
internal/store       SQLite: films, drafts, counted ballots
internal/tmdb        TMDB search with an in-memory cache
internal/web         Gin handlers, htmx fragments, templates, static assets
DESIGN.md            the visual system
render.yaml          Render Blueprint (free web service)
```

## Speed

- Leaderboard is rebuilt off the request path after each vote. Reading it is a pointer load.
- The home page is rendered once per board version and served pre-gzipped: about 8µs per request in `go test -bench Home ./internal/web`.
- CSS is inlined, the font is preloaded, and static files are hashed and cached for a year.
- The home page is about 8.5KB gzipped. There's no JS framework, just htmx (16KB gzipped).
