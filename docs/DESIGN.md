# Filmstop design

## Concept

**Title sequences, cut from paper.** Each page plays like a Saul Bass opening (*Vertigo*, *Anatomy of a Murder*, *The Man with the Golden Arm*): flat fields of colour, shapes cut with scissors, type set big and slightly off its line, and a few seconds of movement before the story starts.

The films still supply the pictures, but every still is printed in its film's colours and cropped by a cut shape, the way Bass cut photographs into his posters.

## Scene

Someone choosing the films that made them, in the evening, on a phone or a laptop, half wanting to show it to a friend. The page should feel like the moment the lights go down: a little theatrical, then out of the way.

The base of the page is the cinema in the dark, a warm black. Colour arrives with the films.

## Colour

### The house

| Token | Value | Use |
| --- | --- | --- |
| `--ink` | `#15110E` | Warm black: the page, and text on light fields |
| `--paper` | `#F4ECDD` | Off-white paper: text on dark fields, primary button |
| `--paper-2` | `#A69F95` | Secondary text on the black page (7.17:1) |
| `--paper-ink-2` | `#55493F` | Secondary text on paper sections and the search bar (7.42:1) |
| `--ink-raised` | `#221C17` | A row lifted off the black page on hover (no text contrast role) |
| `--cut` | `#E2321F` | Vermilion. Shapes only, never text: the brand mark, cut shapes, focus, the active scene |

### Film fields

Each film gets one field, picked from its poster's dominant colour (`tmdb.dominant` measures it; `web.fieldOf` maps it). Every field was solved with `contrast.py` so that its ink reaches **8:1** and its secondary ink **7:1**.

| Field | Colour | Ink | Ink 2 | Shapes | Posters that land here |
| --- | --- | --- | --- | --- | --- |
| `orange` | `#F49A44` | `#15110E` (8.54) | `#302114` (7.06) | vermilion | gold, amber, warm |
| `mustard` | `#EDB32B` | `#15110E` (9.91) | `#3A2D13` (7.08) | vermilion | yellow |
| `rose` | `#E9A6A0` | `#15110E` (9.32) | `#352724` (7.11) | vermilion | pink, purple |
| `aqua` | `#8FC8C0` | `#15110E` (10.01) | `#2B322E` (7.00) | deep teal | pale green and blue |
| `red` | `#7A1810` | `#F4ECDD` (9.10) | `#E3CEC0` (7.05) | orange | red |
| `teal` | `#134A50` | `#F4ECDD` (8.43) | `#DBDACD` (7.03) | mustard | green, cyan |
| `navy` | `#1D2B57` | `#F4ECDD` (11.65) | `#BCBABA` (7.08) | vermilion | blue |
| `ink` | `#2A2420` | `#F4ECDD` (13.04) | `#BDB5AA` (7.55) | vermilion | black, white and grey. Lifted off the black page so it still reads as a panel |

Vivid vermilion can't carry text at 8:1 with either ink, so it stays a shape colour. Fields are set with a class (`.field-orange`), which defines `--field`, `--on`, `--on-2` and `--shape`; components only read those four.

## Type

**Bricolage Grotesque** (variable: weight 200–800, width 75–100%, optical size 12–96), self-hosted, one family for everything.

- **Titles:** weight 800, width 75–85%, tight leading (0.86), set large. Big titles are lowercase-friendly sentence case, never all caps with tracking.
- **Numbers:** places and points are display numerals, weight 800, `font-variant-numeric: tabular-nums` in lists.
- **Body:** weight 420, width 100%, 17–18 px.
- **The cut:** hero titles sit a few degrees off level (`rotate(-2deg)`) on a cut-paper strip, as if pasted down. Only one element per view does this.

Nothing goes below 14 px. Secondary text is at least 7:1, primary at least 8:1.

## Shapes

- **Cut edges:** stills and panels are clipped with `clip-path: polygon(...)` into slightly irregular quadrilaterals, never perfect rectangles and never rounded corners. Three cut shapes (`--cut-a`, `--cut-b`, `--cut-c`) are reused so the page doesn't look random.
- **Printed stills:** photos are greyscale with raised contrast, multiplied onto the field colour (`mix-blend-mode: multiply`), so every still looks printed in its film's ink. Hover or focus reveals full colour.
- **Bars and dots:** the decorative vocabulary is Bass's: thick bars, a solid disc, a jagged strip. No icons that repeat labels, no film-reel or clapperboard clip art.
- **No containers for their own sake.** Sections are separated by changes of field, not by boxes or hairlines.

## Layout

- **The list:** the number one fills the first screen with its field, a printed still in a cut shape, and its title on a pasted strip. Two to five follow as four poster-like tiles, each in its own field. Six to thirty roll as end credits on black.
- **Your five:** the ballot is five paper strips, numbered 1–5, each in its film's field and stacked with slight alternating tilts. Empty slots are dashed outlines marked "cut here". Search sits beside it on desktop and below it on phones.
- **Film pages and cards:** the film's field takes the whole first screen.

## Motion

One authored moment per page, then stillness.

- **Opening (the list, film pages, cards):** on load, two or three bars sweep across and leave the field behind, then the title strip drops into place (about 900 ms total).
- **Picking a film:** the new strip slides into its slot from the side and settles with a small overshoot. Reordering animates the strips into their new places (view transitions, named per film).
- **Casting:** a short title sequence: bars wipe across, your five titles roll as credits, then "That's a wrap" (about 2.8 s; click or any key skips it).
- **Hover:** cut shapes shift a few pixels and stills regain their colour, 200 ms.
- **`prefers-reduced-motion`:** every sequence is replaced by its final frame.

## Refuse

Rounded cards, glass, gradients as decoration, neon, drop shadows (and hard offset "brutalist" shadows), film-reel and clapperboard icons, popcorn, star ratings, uppercase tracked labels, emoji, Inter and every font on the overused list, and more than one tilted element per view.
