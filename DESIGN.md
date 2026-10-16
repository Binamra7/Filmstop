# Filmstop design

## Concept

Opening titles. Each page plays like a film:

- **The list:** opens on the number one film's own scene, a widescreen still with a title
  card and its tagline. Numbers two to five follow as scene selection, and everything
  after rolls as end credits.
- **Your ballot:** a storyboard of five scenes.

The films supply the pictures and the stories. The interface stays out of the frame.

## Scene

Someone choosing the films that made them, on a phone or laptop in the evening.
Cinema is watched in the dark, so the page is dark only, with black letterbox around
the pictures.

## Colour roles

| Token | Value | Use |
| --- | --- | --- |
| `--black` | `#0A0A0A` | Letterbox: the page |
| `--surface` | `#141414` | Raised panels: the storyboard, search box |
| `--surface-2` | `#1C1C1C` | Filled scenes, placeholders |
| `--text` | `#F5F3EF` | Text, primary button fill (15.4:1 or more) |
| `--text-2` | `#B0ACA5` | Secondary text (7.5:1 or more on every surface) |
| `--line` | text at 12% | Hairlines |

There's no accent colour. All colour comes from the stills and posters.

## Type

- **Bodoni Moda** (variable, `opsz` and `wght`): film titles and page headings only,
  the "title card" voice. Upright, weight 500, optical size follows the font size.
- **Mona Sans** (variable): everything you read or press.

Nothing goes below 14px.

## Layout

- **Home:**
  - A full-bleed opening still: 2.39:1 on desktop, 16:9 on phones.
  - The title card sits bottom-left over a scrim on desktop, and below the still on phones.
  - Scene selection is a 2×2 grid of 16:9 stills.
  - End credits are a centred two-column credit roll: place and points on the right of
    the axis, title and director on the left.
- **Ballot:** a sticky storyboard on the left (5 scenes, 16:9 thumbnails), search on
  the right. On phones the storyboard comes first.

## Shape and depth

10px radius on stills, panels and buttons. No shadows on the page except under
floating stills. Scrims are plain gradients to black, only where text sits on
pictures.

## Motion

- **Opening shot:** the hero still settles from 1.06 to 1.0 over 1.6s on load.
- **Scene stills:** they zoom 3% on hover.
- **Page changes:** native view transitions.
- **Reduced motion:** `prefers-reduced-motion` removes the zooms.

## Refuse

Neo-brutalism (thick borders, hard shadows, flat saturated fields), glassmorphism,
neon or gold accents, clapperboards, popcorn, reel icons, star ratings, uppercase
micro-labels, emoji, Inter, italic display serif.
