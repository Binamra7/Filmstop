<div align="center">

<img src="internal/web/static/favicon.svg" width="88" height="88" alt="Filmstop logo: a vermilion disc cut across by a black bar, on warm black" />

# Filmstop

**Everyone names their five favourite films, in order. Filmstop counts every ballot into one list.**

The list plays like a Saul Bass title sequence: each film in a field of colour taken from its poster,<br />
the number one as the opening scene, two to five as scene tiles, and the rest as end credits.

[![Go 1.26](https://img.shields.io/badge/Go-1.26-15110e?style=flat-square&logo=go&logoColor=00ADD8)](https://go.dev)
[![Gin](https://img.shields.io/badge/server-Gin-15110e?style=flat-square&logo=go&logoColor=F4ECDD)](https://gin-gonic.com)
[![htmx 2](https://img.shields.io/badge/htmx-2.0-15110e?style=flat-square&logo=htmx&logoColor=E2321F)](https://htmx.org)
[![SQLite and Turso](https://img.shields.io/badge/data-SQLite_%2F_Turso-15110e?style=flat-square&logo=sqlite&logoColor=8FC8C0)](https://turso.tech)
[![No JS build](https://img.shields.io/badge/JS_build-none-E2321F?style=flat-square&labelColor=15110e)](#how-it-works)
[![CI](https://img.shields.io/github/actions/workflow/status/Binamra7/Filmstop/ci.yml?branch=main&style=flat-square&label=CI&labelColor=15110e)](https://github.com/Binamra7/Filmstop/actions/workflows/ci.yml)

[Live site](https://filmstop.onrender.com) · [Screenshots](#screenshots) · [Features](#features) · [Getting started](#getting-started) · [How it works](#how-it-works) · [Design](#design) · [Docs](#docs)

<br />

<img src="docs/screenshots/00-opening-titles.gif" alt="The opening title sequences on the live site, one after another: bars sweep across and leave each film in its poster colour, then the title drops onto a paper strip. The Shawshank Redemption at number one in orange, Vertigo in red, Inception in teal and The Dark Knight in navy" />

</div>

## The idea

Most "best films" lists are one critic's opinion or an average of star ratings. Filmstop asks a narrower question: which five films made you, in order? Every ballot goes into one shared list, and position counts. A first place earns five points and a fifth place one, so a film that many people put first beats one that many people only mention.

Each page plays like a Saul Bass title sequence cut from paper. Every film gets a flat field of colour taken from its own poster: the server measures the poster's dominant colour and maps it to one of eight fields, and each field's inks were solved to reach 8:1 contrast for primary text and 7:1 for secondary. Stills are printed in the field's colour and cropped by cut shapes, the title is pasted on a strip a few degrees off level, and casting your five plays a short title sequence that rolls your films as credits.

<p align="center">
  <img src="docs/screenshots/10-fields-grid.jpg" alt="Six film pages, each in a colour field taken from its poster: Taxi Driver in mustard, Vertigo in red, Inception in teal, Drive in rose, Back to the Future in navy and Interstellar in orange" />
</p>

<p align="center">
  <img src="https://img.shields.io/badge/orange-F49A44?style=for-the-badge" alt="orange #F49A44" />
  <img src="https://img.shields.io/badge/mustard-EDB32B?style=for-the-badge" alt="mustard #EDB32B" />
  <img src="https://img.shields.io/badge/rose-E9A6A0?style=for-the-badge" alt="rose #E9A6A0" />
  <img src="https://img.shields.io/badge/aqua-8FC8C0?style=for-the-badge" alt="aqua #8FC8C0" /><br />
  <img src="https://img.shields.io/badge/red-7A1810?style=for-the-badge" alt="red #7A1810" />
  <img src="https://img.shields.io/badge/teal-134A50?style=for-the-badge" alt="teal #134A50" />
  <img src="https://img.shields.io/badge/navy-1D2B57?style=for-the-badge" alt="navy #1D2B57" />
  <img src="https://img.shields.io/badge/ink-2A2420?style=for-the-badge" alt="ink #2A2420" />
  <img src="https://img.shields.io/badge/vermilion_(shapes_only)-E2321F?style=for-the-badge" alt="vermilion #E2321F, used for shapes only" />
</p>

There are no accounts. A signed cookie gives each visitor one ballot, which they can change at any time.

## Screenshots

Captured on 17 October 2026 from the [live site](https://filmstop.onrender.com). The list holds seeded demo ballots from the seed tool (made-up voters choosing from real films), so the counts are invented, but the pages are what visitors see.

<table>
  <tr>
    <td width="50%"><img src="docs/screenshots/01-desktop-home.jpg" alt="Desktop: The Shawshank Redemption at number one, its still printed in orange on an orange field beside a vermilion disc, with the title on a tilted paper strip" /></td>
    <td width="50%"><img src="docs/screenshots/02-desktop-ballot.jpg" alt="Desktop: a half-filled storyboard of numbered paper strips, with empty slots marked cut here, beside search results for 'before'" /></td>
  </tr>
  <tr>
    <td align="center"><sub>The list opens on the number one</sub></td>
    <td align="center"><sub>Your five: a storyboard and live search</sub></td>
  </tr>
  <tr>
    <td width="50%"><img src="docs/screenshots/03-desktop-film.jpg" alt="Desktop: Vertigo's film page, Saul Bass's spiral artwork printed on a deep red field, with the title pasted on a cream strip" /></td>
    <td width="50%"><img src="docs/screenshots/04-desktop-your-five.jpg" alt="Desktop: a voter's five card, In the Mood for Love at number one on a red field, its still cut into an uneven shape" /></td>
  </tr>
  <tr>
    <td align="center"><sub>A film page: Vertigo</sub></td>
    <td align="center"><sub>A shareable five</sub></td>
  </tr>
</table>

<table>
  <tr>
    <td width="33%"><img src="docs/screenshots/06a-phone-home.jpg" alt="Phone: The Shawshank Redemption at number one on an orange field" /></td>
    <td width="33%"><img src="docs/screenshots/06b-phone-ballot.jpg" alt="Phone: the ballot with Spirited Away in slot one and an empty slot marked cut here" /></td>
    <td width="33%"><img src="docs/screenshots/06c-phone-your-five.jpg" alt="Phone: a five card with In the Mood for Love at number one, then 2001: A Space Odyssey and Tokyo Story" /></td>
  </tr>
</table>

<details>
<summary><b>More screenshots</b></summary>
<br />

<img src="docs/screenshots/05-desktop-title-sequence.jpg" alt="Three frames of the title sequence after casting: red and cream bars wipe across the ballot, the five chosen films roll up as credits, then 'That's a wrap.' lands on a paper strip" />
<p align="center"><sub>Casting plays a title sequence of your five (frames from a local build)</sub></p>

<img src="docs/screenshots/07-desktop-two-to-five.jpg" alt="Desktop: places two to five as four tiles, each in its own colour field: The Dark Knight, Pulp Fiction, The Godfather and Spirited Away" />
<p align="center"><sub>Two to five: scene tiles</sub></p>

<img src="docs/screenshots/08-desktop-credits.jpg" alt="Desktop: Also starring, places six to thirty listed like end credits on black, with points, director, year and running time" />
<p align="center"><sub>End credits: places six to thirty</sub></p>

<img src="docs/screenshots/09-desktop-method.jpg" alt="Desktop: How the list is counted, beside a vermilion disc cut by a black bar, with a bar chart of points per place" />
<p align="center"><sub>How it works: the scoring, explained on the site</sub></p>

</details>

## Features

**The list** (`/`)

- The number one opens full-bleed in its poster's colour field, with its tagline, its points and a bar for each place it was ranked
- Two to five as four scene tiles, each in its own field; six to thirty as rolling end credits on black
- Programme notes written from the votes: photo finishes, cult favourites, the most divisive film, a director with a clear lead
- The latest ballots as they come in, with no voter named

**Your five** (`/ballot`)

- Five numbered paper strips, each in its film's field. Search TMDB, then add, reorder or remove films; the draft saves on every change
- Each pick shows how many others chose it and where it stands on the list
- Casting counts the ballot; casting again later replaces it

**Film pages and cards**

- `/film/:id`: the story, where the film stands, how voters place it, and the films that share its ballots
- `/five/:code`: a shareable page of someone's five with a reading of their taste: which decades, how close to the consensus, kindred ballots

**Motion**

- The list, film pages and cards open with bars sweeping across the field, then the title strip drops into place
- A new pick slides into its slot, a removed one slides out, and moved strips animate to their new places (view transitions named per film, through htmx's `transition:true`)
- Casting plays a short title sequence: bars wipe across, your five titles roll as credits, then "That's a wrap." A click or any key skips it
- Printed stills regain their full colour on hover or keyboard focus
- With `prefers-reduced-motion` set, animations and transitions are turned off and the title sequence is not shown

**Small and fast**

- One Go binary with the templates, CSS, fonts and htmx embedded, and no JavaScript build step
- The list is rebuilt off the request path after each vote. The home page is rendered once per list version and served pre-gzipped
- Every action works without JavaScript; htmx swaps in fragments when it's available

## Getting started

You need Go 1.26 or newer. Film search needs a [TMDB API key](https://www.themoviedb.org/settings/api), either a v3 key or a v4 read token. Without one the site still runs, with search turned off.

```bash
git clone https://github.com/Binamra7/Filmstop.git
cd Filmstop
export TMDB_API_KEY=your_key_here
go run .                                   # http://localhost:8080
```

To fill the list with made-up ballots for development:

```bash
go run ./cmd/seed -db filmstop.db -voters 400
```

The seed needs no network. It casts ballots over a built-in catalogue of 72 real films and prints cookies for five personas (a voter with a card, a kindred ballot, an open draft, a newcomer and more), so you can browse as each one. With `TMDB_API_KEY` set it also fetches the films' real details, stills and posters. Running it again replaces the seeded ballots.

<details>
<summary><b>Configuration</b></summary>
<br />

| Variable | Default | What it does |
| --- | --- | --- |
| `PORT` | `8080` | Port to listen on |
| `DATABASE_PATH` | `filmstop.db` | SQLite file. Put it on a persistent disk |
| `TMDB_API_KEY` | none | Film search and details |
| `TURSO_DATABASE_URL` | none | Use this Turso database instead of the SQLite file |
| `TURSO_AUTH_TOKEN` | none | Token for `TURSO_DATABASE_URL` |

</details>

<details>
<summary><b>Docker</b></summary>
<br />

```bash
cp .env.example .env                       # add your TMDB_API_KEY
docker compose up -d --build               # http://localhost:8080
docker compose run --rm seed && docker compose restart web   # optional demo data
```

The database lives in the `filmstop-data` volume and survives rebuilds; `docker compose down -v` deletes it. The server reads the list at startup, so restart it after seeding. The image is a static binary on distroless and runs as a non-root user.

</details>

<details>
<summary><b>Deploying</b></summary>
<br />

The live site runs on Render's free web service, with the database on Turso's free plan. `render.yaml` describes the service, and Render asks for `TMDB_API_KEY`, `TURSO_DATABASE_URL` and `TURSO_AUTH_TOKEN` when you create it from the Blueprint. [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) has the full steps, including how to seed or reset the production list.

</details>

## How it works

The server is Gin with Go's `html/template`. Pages are plain HTML forms that htmx upgrades: adding, moving or removing a film posts to the server, which answers with the new storyboard fragment. Without JavaScript the same forms submit normally and redirect back.

| Route | Method | What it does |
| --- | --- | --- |
| `/` | GET | The list |
| `/ballot` | GET | Your storyboard and search |
| `/search?q=` | GET | Search results fragment (TMDB, cached for 30 minutes) |
| `/ballot/pick`, `/ballot/move`, `/ballot/remove` | POST | Edit the draft; returns the storyboard fragment |
| `/ballot/cast` | POST | Count the draft as your ballot |
| `/film/:id` | GET | A film's page |
| `/five`, `/five/:code` | GET | Your card, and a shared card by its code |
| `/method` | GET | How the list is counted |
| `/healthz` | GET | Health check |

**Counting.** Each ballot gives 5 points to its first film, down to 1 point for its fifth. Points are summed; ties go to the film with more first places, then more ballots. Plackett–Luce and Condorcet methods were tested and rejected. With top-five ballots drawn from every film ever made, they learn almost nothing from order; `internal/rank/pl_test.go` demonstrates it.

**Identity.** Each visitor gets a random id in an HMAC-signed, HttpOnly cookie. One cookie holds one ballot, and casting again replaces it. A POST whose `Origin` names another site is rejected.

**Colour fields.** When the server fetches a film's details from TMDB, it also downloads the poster at TMDB's smallest size (`w92`, 92 px wide) and measures its dominant colour (`internal/tmdb/swatch.go`): each vivid pixel votes for one of 36 hue bins, weighted by saturation times value, and the winning bin's average colour is stored in the film's `swatch` column. At render time `fieldOf` in `internal/web/field.go` maps that colour's hue to one of the eight fields, with grey and very dark posters going to `ink`. A film not measured yet gets a field from its id, so its colour stays stable.

**Storage.** SQLite through `database/sql`: the pure-Go `modernc.org/sqlite` driver for a local file, or Turso (hosted SQLite) over HTTP in production. Neither needs CGO.

```
main.go              server entry point: config, graceful shutdown, health-check flag
cmd/seed/            demo data: built-in film catalogue, taste groups, personas
internal/rank/       scoring and ordering (pure functions, tested)
internal/store/      ballots, drafts, films and share codes, in SQLite or Turso
internal/tmdb/       TMDB search with an in-memory cache, per-film details, poster colour
internal/web/        Gin handlers, the list keeper, programme notes, colour fields, templates, static files
docs/                design, architecture, decisions, deployment, testing
```

## Design

The full system lives in [`docs/DESIGN.md`](docs/DESIGN.md), and the tokens are in [`internal/web/static/site.css`](internal/web/static/site.css). In short:

- **Fields from posters:** each film takes one of eight colour fields from its poster's dominant colour. Every field defines its own ink, secondary ink and shape colour, checked to 8:1 and 7:1 contrast. Vermilion can't carry text at that contrast, so it is used only for shapes.
- **Type:** one variable family, [Bricolage Grotesque](https://github.com/ateliertriay/bricolage), self-hosted, using its weight, width and optical size axes: heavy and narrow for titles and numbers, regular width for body text. Nothing is smaller than 14 px.
- **Shapes:** stills and panels have cut edges from three reused polygons, stills are printed greyscale onto the field colour, and the decoration is limited to bars, discs and jagged strips. Sections are separated by changes of field, not boxes.
- **Motion:** one authored moment per page (the opening sweep, a pick sliding in, the cast title sequence), then stillness.

## Docs

| Doc | What's in it |
| --- | --- |
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | How a request flows, how the list is kept, the data model |
| [`docs/DECISIONS.md`](docs/DECISIONS.md) | Why Go and htmx, the scoring, SQLite and Turso, the hosting |
| [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) | Render and Turso setup, secrets, seeding and resetting production |
| [`docs/TESTING.md`](docs/TESTING.md) | Tests, benchmarks, CI and manual checks |
| [`docs/DESIGN.md`](docs/DESIGN.md) | The visual system: colour fields, type, shapes, layout, motion |
| [`AGENTS.md`](AGENTS.md) | Working notes for coding agents and people |

## Testing

```bash
go test -race ./...
```

The tests cover the scoring and its tie-breaks, the ballot flow end to end through the HTTP handlers, drafts and casting, share codes, programme notes, taste readings, film pages, and the retry for dropped Turso connections. CI also runs `gofmt`, `go vet`, `govulncheck` and a Docker build on every push and pull request. See [`docs/TESTING.md`](docs/TESTING.md).

## Acknowledgements

- Film data, stills and posters from [TMDB](https://www.themoviedb.org/). This product uses the TMDB API but is not endorsed or certified by TMDB.
- [Bricolage Grotesque](https://github.com/ateliertriay/bricolage) by Mathieu Triay, under the SIL Open Font License (see `internal/web/static/OFL.txt`)
- [htmx](https://htmx.org), under the Zero-Clause BSD licence
- [Gin](https://gin-gonic.com), [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) and [libsql-client-go](https://github.com/tursodatabase/libsql-client-go)
- Visual approach after Saul Bass's title sequences and posters (*Vertigo*, *Anatomy of a Murder*, *The Man with the Golden Arm*)
