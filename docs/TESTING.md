# Testing

| Layer | Tool | Where |
| --- | --- | --- |
| Scoring and order | `go test` | `internal/rank/*_test.go` |
| Store: drafts, casting, tallies, share codes, notes | `go test` against a temporary SQLite file | `internal/store/*_test.go` |
| Turso connection retry | `go test` with a fake transport | `internal/store/remote_test.go` |
| Poster colour measurement | `go test` on generated images | `internal/tmdb/swatch_test.go` |
| HTTP flows, pages, notes, taste readings | `go test` with `httptest` | `internal/web/*_test.go` |
| Swatch to colour field | `go test` | `internal/web/field_test.go` |
| Home page speed | `go test -bench` | `internal/web/bench_test.go` |
| Format and static checks | `gofmt`, `go vet` | CI |
| Known vulnerabilities | `govulncheck` | CI |
| Container | `docker build` | CI |

## Running

```bash
go test -race ./...                       # everything, with the race detector
go test -run TestBallotFlow -v ./internal/web
go test -bench Home -benchmem ./internal/web
gofmt -l . && go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## What the tests check

- **Scoring** (`rank`): points per position, that order beats breadth when it should, every tie-break, and that Plackett–Luce ignores position (the reason it isn't used).
- **Ballot flow** (`web/web_test.go`): pick, move, remove and cast through the real handlers, with and without the `HX-Request` header.
- **Drafts** (`store`): a cleared draft stays cleared instead of bringing the counted ballot back.
- **Film pages and cards:** standings, companions, share codes, the five-card routes.
- **Programme notes:** quiet on small lists, photo-finish ties, a director note only with a clear lead, and arrivals that never name a voter.
- **Poster colour** (`tmdb`): `TestDominant` paints synthetic posters and checks that a small vivid area beats a larger dark one, the larger of two vivid colours wins, and a grey poster returns its unsaturated average. `TestHSV` checks the hue conversion for the primaries and yellow.
- **Colour fields** (`web`): `TestFieldOf` maps swatches measured from real TMDB posters (Vertigo, Amélie, The Godfather and others) and synthetic ones to the expected field, including `ink` for black and grey posters. `TestFieldOfUnmeasured` checks that an empty, `none` or malformed swatch gets a stable non-ink field chosen from the id, and that neighbouring ids differ.
- **Turso retry:** a request that failed on a stale reused connection is resent once with the same body; fresh connections and other errors aren't retried.

Store tests use a real SQLite file in a temporary directory, not mocks.

## CI

`.github/workflows/ci.yml` runs on every push to `main` and every pull request, in three jobs:

1. **Lint and test:** `gofmt`, `go vet`, `go test -race`, `go build`.
2. **Vulnerability scan:** `govulncheck`, so a vulnerable dependency or Go release fails the build.
3. **Docker image:** builds the production image.

Dependabot opens weekly pull requests for Go modules, GitHub Actions and the Docker base images.

## Manual checks

Before a release with visible changes:

- Browse as each seed persona (the seed prints their cookies) to see every state: a counted ballot with a card, an open draft, a half-filled newcomer.
- Check the pages at a phone width (390 px) for horizontal scrolling.
- Run an accessibility checker (axe) on `/`, `/ballot`, a film page and a card. Wait for the opening sequence to finish first, or contrast is measured mid-animation.

Design checks, for any change to colours, type or motion (see [`DESIGN.md`](DESIGN.md)):

- **Contrast:** check every field's inks with `python3 ~/.claude/skills/human-design/scripts/contrast.py check FG BG [FG BG ...]` (or `table palette.json` for the whole set). Primary ink must reach 8:1 on its field and secondary ink 7:1. Vermilion is never used for text.
- **Text size:** nothing renders below 14 px. Search the CSS for `font-size` values and check the smallest text at a phone width.
- **Reduced motion:** turn on reduced motion (DevTools rendering panel, or the OS setting) and check that every page shows its final frame straight away and casting shows no title sequence.
- **View transitions:** in Chromium, page changes cross-fade and ballot strips slide in, out and between places. In Firefox or Safari without view transition support, the same actions must still work, just without the animation.
- **Cast sequence:** after casting, a click or any key removes the sequence, and it removes itself after about 3 seconds.
- On the live site, load a page after it has been idle for a few minutes and check Render's logs for errors.
