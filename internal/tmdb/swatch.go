package tmdb

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg" // TMDB serves posters as JPEG
	"io"
	"math"
	"net/http"
)

// swatch measures the dominant colour of a poster from TMDB's smallest size
// (92 px wide, a few KB). It returns "none" when there's no poster or it can't
// be read, so a film isn't measured again on every rebuild.
func (c *Client) swatch(ctx context.Context, posterPath string) string {
	if posterPath == "" {
		return "none"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://image.tmdb.org/t/p/w92"+posterPath, nil)
	if err != nil {
		return "none"
	}
	res, err := c.http.Do(req)
	if err != nil {
		return "none"
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "none"
	}
	img, _, err := image.Decode(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "none"
	}
	return dominant(img)
}

// dominant returns the colour that most of an image's vivid pixels share, as
// "#rrggbb". Each pixel votes for one of 36 hue bins with a weight of
// saturation × value, so a small saturated title beats a large grey sky. An
// image with almost no vivid pixels (a black-and-white poster) returns its
// average colour, which is unsaturated.
func dominant(img image.Image) string {
	const bins = 36
	var weight [bins]float64
	var sum [bins][3]float64
	var avg [3]float64
	var vivid, n float64
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r16, g16, b16, _ := img.At(x, y).RGBA()
			r, g, bl := float64(r16>>8), float64(g16>>8), float64(b16>>8)
			avg[0], avg[1], avg[2] = avg[0]+r, avg[1]+g, avg[2]+bl
			n++
			h, s, v := hsv(r, g, bl)
			w := s * v
			if w < 0.12 {
				continue
			}
			i := int(h/360*bins) % bins
			weight[i] += w
			sum[i][0], sum[i][1], sum[i][2] = sum[i][0]+r*w, sum[i][1]+g*w, sum[i][2]+bl*w
			vivid += w
		}
	}
	if n == 0 {
		return "none"
	}
	if vivid/n < 0.04 {
		return hex(avg[0]/n, avg[1]/n, avg[2]/n)
	}
	// Score each bin with its neighbours, so a hue split across a bin edge
	// still wins.
	best, bestScore := 0, -1.0
	for i := range bins {
		score := weight[(i+bins-1)%bins] + 2*weight[i] + weight[(i+1)%bins]
		if score > bestScore {
			best, bestScore = i, score
		}
	}
	w := weight[best]
	if w == 0 {
		return hex(avg[0]/n, avg[1]/n, avg[2]/n)
	}
	return hex(sum[best][0]/w, sum[best][1]/w, sum[best][2]/w)
}

// hsv converts 0–255 RGB to hue in degrees and saturation and value in 0–1.
func hsv(r, g, b float64) (h, s, v float64) {
	r, g, b = r/255, g/255, b/255
	hi, lo := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	v = hi
	if hi == 0 {
		return 0, 0, 0
	}
	d := hi - lo
	s = d / hi
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
	h *= 60
	if h < 0 {
		h += 360
	}
	return h, s, v
}

func hex(r, g, b float64) string {
	return fmt.Sprintf("#%02x%02x%02x", uint8(math.Round(r)), uint8(math.Round(g)), uint8(math.Round(b)))
}
