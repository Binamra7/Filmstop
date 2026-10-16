# AGENTS.md

Notes for coding agents (and people) working in this repo. `CLAUDE.md` imports this file, so keep everything here.

## What this is

Filmstop: visitors name their five favourite films in order, and every ballot is counted into one list. Go 1.26, Gin, `html/template` and htmx 2, all served from one binary. The data is in SQLite locally and Turso in production. Live at https://filmstop.onrender.com.

## Production guardrails

- **`main` is production.** Render deploys every push to `main`. Work on a branch, open a pull request, and merge only after CI passes and the owner agrees.
- **Production data is on Turso.** The credentials live in Render and in `~/.config/filmstop/turso.env` on the owner's machine. Never put them in `.env`, in code, or in a commit.
- **Don't seed production by accident.** `cmd/seed` refuses a remote database unless you pass `-remote`. Production already holds about 5,000 seeded ballots on purpose (see `docs/DEPLOYMENT.md`).
- **After any direct database change, restart the Render service.** The list is built in memory at startup.
- **Secrets never appear in logs or errors.** TMDB errors already drop the request URL because it can carry the key; keep it that way.

## Commands

| Task | Command |
| --- | --- |
| Run locally (localhost:8080) | `TMDB_API_KEY=... go run .` |
| Seed a local database | `go run ./cmd/seed -db filmstop.db -voters 400` |
| All tests, race detector | `go test -race ./...` |
| Format and vet (CI fails otherwise) | `gofmt -l . && go vet ./...` |
| Vulnerability scan | `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` |
| Home page benchmark | `go test -bench Home ./internal/web` |
| Docker | `docker compose up -d --build` |

## Where things live

```
main.go                 entry point: env config, graceful shutdown, -healthcheck flag
cmd/seed/               demo data
  films.go              the built-in catalogue of 72 films
  main.go               taste groups, personas, the -remote guard
internal/rank/          Points and Order; pure, no I/O
internal/store/
  store.go              schema, migrations, Open/Source, drafts, casting, tallies
  five.go               share codes and kindred ballots for cards
  film.go               a film's tally and companion films
  notes.go              latest ballots; ballot_notes.go: how many share a pick
  remote.go             retrying HTTP transport for Turso
internal/tmdb/          search (30-minute cache) and per-film details
internal/web/
  server.go             routes, templates, page cache, static assets
  handlers.go           list, ballot, search, pick/move/remove/cast
  board.go              boardKeeper: the in-memory list, rebuilt after casts
  notes.go              programme notes and latest ballots
  film.go, five.go      film pages, cards and taste readings
  voter.go              signed voter cookie, same-origin check
  templates/, static/   embedded with go:embed
docs/                   ARCHITECTURE, DECISIONS, DEPLOYMENT, TESTING, DESIGN
render.yaml             Render Blueprint
```

## Conventions

- **Go style:** `gofmt`, standard library first, no new dependencies without a reason in `docs/DECISIONS.md`. The build must stay CGO-free (`CGO_ENABLED=0`).
- **Comments:** every package and exported identifier has a doc comment that starts with its name. Elsewhere, comment only to explain why: a concurrency choice, a SQL subtlety, a security reason. Write plainly; no filler.
- **Handlers:** every action must work without JavaScript. Answer htmx requests (`HX-Request` header) with a fragment, and redirect plain form posts.
- **Store:** multi-statement changes go in one transaction. The SQL must run on both SQLite and Turso (libSQL), so avoid driver-specific pragmas outside the local DSN.
- **The list:** never compute the board per request. Change what `boardKeeper.rebuild` produces, and call `Touch` after anything that changes ballots.
- **Templates and CSS:** follow `docs/DESIGN.md`. Dark only, no accent colour, nothing smaller than 14px, secondary text at 7:1 contrast or more. CSS custom properties live in `static/site.css`.
- **Tests:** next to the code (`foo_test.go`). Store tests use a real SQLite file in `t.TempDir()`, not mocks.
- **Commits:** Conventional Commits, title line only (`fix(store): ...`, `docs: ...`).

## Common tasks

- **Change the scoring:** `internal/rank`, then update the Method page template (`templates/method.html`) and `docs/DECISIONS.md`.
- **Add a page:** a handler and route in `internal/web`, a template in `templates/`, and a test in `internal/web`.
- **Add a column:** append it to `migrations` in `store/store.go`; existing databases get it on the next start.
- **Deploy or troubleshoot production:** `docs/DEPLOYMENT.md`.
- **Why is X done this way?** `docs/DECISIONS.md`.
