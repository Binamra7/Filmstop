package web

import "testing"

func TestFieldOf(t *testing.T) {
	// Swatches measured from real TMDB posters.
	for _, c := range []struct {
		film, swatch, want string
	}{
		{"The Shawshank Redemption", "#dca45c", "orange"},
		{"Vertigo", "#de5132", "red"},
		{"In the Mood for Love", "#b81a23", "red"},
		{"Amélie", "#027936", "teal"},
		{"The Godfather", "#211b1a", "ink"},
		{"Tokyo Story", "#6d6d6d", "ink"},
		{"a yellow poster", "#e8c22a", "mustard"},
		{"a pale blue-green poster", "#8fc8c0", "aqua"},
		{"a blue poster", "#2b4c9b", "navy"},
		{"a pink poster", "#d0609a", "rose"},
	} {
		if got := fieldOf(c.swatch, 1); got != c.want {
			t.Errorf("%s (%s): got %s, want %s", c.film, c.swatch, got, c.want)
		}
	}
}

func TestFieldOfUnmeasured(t *testing.T) {
	for _, swatch := range []string{"", "none", "#zzzzzz"} {
		if a, b := fieldOf(swatch, 42), fieldOf(swatch, 42); a != b || a == "ink" {
			t.Errorf("unmeasured %q: got %s then %s; want one stable colour field", swatch, a, b)
		}
	}
	if fieldOf("", 1) == fieldOf("", 2) {
		t.Error("neighbouring ids should get different fields")
	}
}
