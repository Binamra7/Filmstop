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

### Saul Bass title sequences, coloured by the posters
The first design was dark only with no accent colour. It was readable but quiet, and the owner wanted the site to feel fun and interactive. The redesign treats each page as a Saul Bass opening: flat colour fields, cut-paper shapes, big type pasted slightly off level, and a short burst of movement. [`DESIGN.md`](DESIGN.md) has the full system.

- **Fields come from the posters.** Each film's panel takes its colour from its own poster, so a film looks like itself across the list, the ballot and its page, and the colour arrives with the films instead of being imposed on them. The page underneath stays a warm black.
- **A fixed set of fields, not raw poster colours.** A poster's dominant colour can be anything, and text on an arbitrary colour can't be guaranteed to read. So the measured colour is only used to choose one of eight fields, each with inks solved in advance to 8:1 for primary text and 7:1 for secondary. Any poster gives a readable panel, and the CSS stays a handful of classes. Trade-off: two posters of slightly different reds get the same red.
- **Vermilion is for shapes only.** The house accent, vermilion `#E2321F`, can't reach 8:1 with either the black or the paper ink, so it never carries text. It marks the brand disc, cut shapes, focus and the active scene.
- **Bricolage Grotesque, one variable family.** It replaced two fonts (Mona Sans and Bodoni Moda). Its weight (200–800) and width (75–100%) axes give the narrow, heavy display titles and the plain body text from the same files, so there's one family to load and pair. It isn't on the list of overused fonts the design refuses.
- **CSS-only motion, one moment per page.** The openings, the strip movements and the cast sequence are CSS animations and view transitions, so there's still no JavaScript build, they survive htmx swaps, and `prefers-reduced-motion` turns them all off in one media query. Each page gets one authored moment and then stays still, because a sequence that repeats on every interaction stops being fun and starts getting in the way. Trade-off: view transitions between pages and on the ballot only animate in browsers that support them (Chromium); others just swap.

### The seed refuses remote databases by default
The seed's ballots are made up. Production is a Turso URL, and that URL may be set in a shell for other reasons, so the seed refuses any remote database unless `-remote` is passed. Production was seeded on purpose for the demo; see [`DEPLOYMENT.md`](DEPLOYMENT.md).
