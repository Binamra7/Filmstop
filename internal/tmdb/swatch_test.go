package tmdb

import (
	"image"
	"image/color"
	"testing"
)

// poster fills a 92×138 image with bg, then paints the top fraction with fg.
func poster(bg, fg color.RGBA, fraction float64) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 92, 138))
	for y := range 138 {
		for x := range 92 {
			c := bg
			if float64(y) < 138*fraction {
				c = fg
			}
			img.Set(x, y, c)
		}
	}
	return img
}

func TestDominant(t *testing.T) {
	black := color.RGBA{12, 12, 12, 255}
	grey := color.RGBA{128, 128, 128, 255}
	red := color.RGBA{220, 40, 30, 255}
	blue := color.RGBA{30, 60, 200, 255}
	cases := []struct {
		name string
		img  image.Image
		want string
	}{
		{"a vivid colour beats a larger dark field", poster(black, red, 0.2), "#dc281e"},
		{"the larger of two vivid colours wins", poster(blue, red, 0.3), "#1e3cc8"},
		{"a grey poster returns its unsaturated average", poster(grey, grey, 0), "#808080"},
	}
	for _, c := range cases {
		if got := dominant(c.img); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}

func TestHSV(t *testing.T) {
	for _, c := range []struct {
		r, g, b float64
		h       float64
	}{{255, 0, 0, 0}, {0, 255, 0, 120}, {0, 0, 255, 240}, {255, 255, 0, 60}} {
		if h, s, v := hsv(c.r, c.g, c.b); h != c.h || s != 1 || v != 1 {
			t.Errorf("hsv(%v,%v,%v) = %v,%v,%v; want hue %v", c.r, c.g, c.b, h, s, v, c.h)
		}
	}
}
