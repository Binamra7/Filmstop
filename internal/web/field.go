package web

import (
	"math"
	"strconv"
)

// fields are the poster colours a film's panel can take, named after the CSS
// classes in static/site.css (".field-orange" and so on). Each one has inks
// checked to 8:1 for primary text and 7:1 for secondary (docs/DESIGN.md), so any
// poster maps to a readable panel.
var fields = []string{"orange", "mustard", "rose", "aqua", "red", "teal", "navy"}

// fieldOf picks the colour field for a film from its poster's measured swatch.
// Grey and near-black posters get the ink field. A film not measured yet gets a
// field chosen by its id, so it keeps the same colour until it is measured.
func fieldOf(swatch string, id int64) string {
	r, g, b, ok := parseHex(swatch)
	if !ok {
		return fields[int(uint64(id)%uint64(len(fields)))]
	}
	h, s, v := hsv(r, g, b)
	switch {
	case s < 0.25 || v < 0.25:
		return "ink"
	case h < 20 || h >= 345:
		return "red"
	case h < 42:
		return "orange"
	case h < 70:
		return "mustard"
	case h < 200:
		if v > 0.6 && s < 0.6 {
			return "aqua"
		}
		return "teal"
	case h < 260:
		return "navy"
	default:
		return "rose"
	}
}

func parseHex(s string) (r, g, b float64, ok bool) {
	if len(s) != 7 || s[0] != '#' {
		return 0, 0, 0, false
	}
	n, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	return float64(n >> 16 & 0xff), float64(n >> 8 & 0xff), float64(n & 0xff), true
}

// hsv converts 0–255 RGB to hue in degrees and saturation and value in 0–1.
func hsv(r, g, b float64) (h, s, v float64) {
	r, g, b = r/255, g/255, b/255
	hi, lo := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	if hi == 0 {
		return 0, 0, 0
	}
	d := hi - lo
	s, v = d/hi, hi
	if d == 0 {
		return 0, s, v
	}
	switch hi {
	case r:
		h = math.Mod((g-b)/d, 6)
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	if h *= 60; h < 0 {
		h += 360
	}
	return h, s, v
}
