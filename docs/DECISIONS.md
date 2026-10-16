# Decisions

### Go, Gin and htmx instead of React and Express
The first version was a React frontend with an Express API, and it was never finished. The site is mostly lists and forms, which server-rendered HTML does well. Go gives one small binary with everything embedded; htmx adds partial updates (search as you type, editing the storyboard) without a JavaScript build. Trade-off: richer client-side interactions, like drag-to-reorder, would need hand-written JavaScript.

### Positional scoring instead of Plackett–Luce or Condorcet
A ballot is a top five chosen from every film ever made. Models that treat unlisted films as "ranked below" learn almost everything from whether a film was listed and almost nothing from where: a film always ranked first and one always ranked fifth by the same number of people end up within a few percent of each other. Filmstop's premise is that order matters, so it uses explicit weights (5, 4, 3, 2, 1). `internal/rank/pl_test.go` demonstrates the problem, and the site's Method page explains it to visitors.

### No accounts
Signing up is a bigger ask than naming five films. A random id in an HMAC-signed cookie gives each browser one ballot it can change. Trade-off: one person can vote from several browsers. For a list of favourite films that's acceptable; it would not be for anything with stakes.

### The list lives in memory
Every page shows the list, and it changes only when someone casts. So it's rebuilt in a background loop after each cast and read with an atomic pointer load, and the home page is rendered once per list version. Trade-off: a change written straight to the database needs a server restart to show.

### SQLite, and Turso in production
SQLite needs no separate server and fits the data, which is a few small tables. Locally it's a file through `modernc.org/sqlite`, which is pure Go, so builds need no C compiler. Production uses Turso, hosted SQLite, because the free host wipes its disk on every restart. Both run the same SQL through `database/sql`; `store.Source` picks one from the environment.

### `libsql-client-go`, though it's deprecated
Turso's recommended Go drivers (`go-libsql`, `turso-go`) need CGO or native libraries, which would break the static, distroless image. `libsql-client-go` is pure Go and still receives fixes. It sends statements over HTTP through `http.DefaultClient` and can't be given its own client, which is why `store/remote.go` installs a retrying transport there. If a pure-Go driver from Turso appears, switch to it.

### Render and Turso free plans
The host had to be free with no time limit. Vercel runs Go only as short-lived functions, so the in-memory list and SQLite file don't fit. Fly.io's free use ends after a trial. Render's free web service runs the Dockerfile with no time limit but sleeps after 15 idle minutes and has no persistent disk, so the data goes to Turso, whose free plan is far more than this site needs. Trade-off: the first visit after a quiet spell waits about a minute for the server to wake.

### Distroless, non-root image
The binary is static (`CGO_ENABLED=0`), so the runtime image needs no shell or libc. `gcr.io/distroless/static-debian12:nonroot` keeps the image small and runs as uid 65532. The container health check is the binary itself (`/filmstop -healthcheck`), since there's no `curl` in the image.

### Dark only, no accent colour
Films are watched in the dark, and the stills and posters already bring colour. A second theme or an accent would compete with them. [`DESIGN.md`](DESIGN.md) has the full system and the list of things it refuses.

### The seed refuses remote databases by default
The seed's ballots are made up. Production is a Turso URL, and that URL may be set in a shell for other reasons, so the seed refuses any remote database unless `-remote` is passed. Production was seeded on purpose for the demo; see [`DEPLOYMENT.md`](DEPLOYMENT.md).
