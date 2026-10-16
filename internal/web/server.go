// Package web serves Filmstop's pages and htmx fragments.
package web

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Binamra7/filmstop/internal/rank"
	"github.com/Binamra7/filmstop/internal/store"
	"github.com/Binamra7/filmstop/internal/tmdb"
	"github.com/gin-gonic/gin"
)

//go:embed templates static
var assets embed.FS

// Server holds what the handlers share: the store, the TMDB client, the
// current board and the parsed templates and static files.
type Server struct {
	st       *store.Store
	tmdb     *tmdb.Client
	board    *boardKeeper
	voters   voters
	pages    map[string]*template.Template
	static   map[string]asset
	homePage homeCache
	// ctx ends at shutdown; background work started by a request hangs off it.
	ctx context.Context
}

// New builds the first board and loads the embedded templates and static
// files. Cancelling ctx stops the board's background rebuilds and any detail
// fetches a request started.
func New(ctx context.Context, st *store.Store, tm *tmdb.Client) (*Server, error) {
	secret, err := st.Secret(ctx)
	if err != nil {
		return nil, err
	}
	board, err := newBoardKeeper(ctx, st, tm)
	if err != nil {
		return nil, err
	}
	s := &Server{st: st, tmdb: tm, board: board, voters: voters{secret: secret}, ctx: ctx}
	if s.static, err = loadStatic(); err != nil {
		return nil, err
	}
	if s.pages, err = s.loadPages(); err != nil {
		return nil, err
	}
	return s, nil
}

// Handler returns the site's routes, /healthz included.
func (s *Server) Handler() http.Handler {
	r := gin.New()
	r.Use(gin.Recovery(), sameOrigin)
	r.GET("/static/:name", s.serveStatic)
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	v := r.Group("/", s.voters.middleware)
	v.GET("/", s.home)
	v.GET("/method", s.method)
	v.GET("/ballot", s.ballotPage)
	v.GET("/search", s.search)
	v.POST("/ballot/pick", s.pick)
	v.POST("/ballot/move", s.move)
	v.POST("/ballot/remove", s.remove)
	v.POST("/ballot/cast", s.cast)
	// Feature routes live next to their handlers.
	s.filmRoutes(v)
	s.fiveRoutes(v)
	r.NoRoute(func(c *gin.Context) { s.render(c, http.StatusNotFound, "notfound", page{"Not found · Filmstop", ""}) })
	return r
}

// --- templates

type spreadBar struct{ Pos, N, Pct int }

var funcs = template.FuncMap{
	"poster": func(p, size string) string {
		if p == "" {
			return ""
		}
		return "https://image.tmdb.org/t/p/" + size + p
	},
	"num": func(n int) string {
		s := fmt.Sprint(n)
		for i := len(s) - 3; i > 0; i -= 3 {
			s = s[:i] + "," + s[i:]
		}
		return s
	},
	"plural": func(n int, one, many string) string {
		if n == 1 {
			return one
		}
		return many
	},
	"ago": func(t time.Time) string {
		d := time.Since(t)
		switch {
		case d < time.Minute:
			return "just now"
		case d < time.Hour:
			return fmt.Sprintf("%d min ago", int(d.Minutes()))
		case d < 48*time.Hour:
			return fmt.Sprintf("%d h ago", int(d.Hours()))
		}
		return t.Format("2 Jan 2006")
	},
	"avg": func(f float64) string { return fmt.Sprintf("%.1f", f) },
	"add": func(a, b int) int { return a + b },
	"runtime": func(min int) string {
		switch {
		case min <= 0:
			return ""
		case min < 60:
			return fmt.Sprintf("%d\u00a0min", min)
		case min%60 == 0:
			return fmt.Sprintf("%d\u00a0h", min/60)
		}
		// Non-breaking spaces keep "2 h 37 min" on one line.
		return fmt.Sprintf("%d\u00a0h\u00a0%d\u00a0min", min/60, min%60)
	},
	"sub": func(a, b int) int { return a - b },
	// spread turns per-position counts into bar heights relative to the largest.
	"spread": func(by [rank.Slots]int) []spreadBar {
		top := 1
		for _, n := range by {
			top = max(top, n)
		}
		out := make([]spreadBar, len(by))
		for i, n := range by {
			out[i] = spreadBar{Pos: i + 1, N: n, Pct: max(4, n*100/top)}
		}
		return out
	},
	"ordinal": func(n int) string {
		return map[int]string{1: "first", 2: "second", 3: "third", 4: "fourth", 5: "fifth"}[n]
	},
}

// loadCSS inlines site.css first, then every other static/*.css in name order,
// so a feature can keep its styles in its own file.
func loadCSS() (string, error) {
	names, err := fs.Glob(assets, "static/*.css")
	if err != nil {
		return "", err
	}
	sort.Slice(names, func(i, j int) bool {
		if (names[i] == "static/site.css") != (names[j] == "static/site.css") {
			return names[i] == "static/site.css"
		}
		return names[i] < names[j]
	})
	var b strings.Builder
	for _, n := range names {
		css, err := assets.ReadFile(n)
		if err != nil {
			return "", err
		}
		b.Write(css)
		b.WriteByte('\n')
	}
	return b.String(), nil
}

func (s *Server) loadPages() (map[string]*template.Template, error) {
	css, err := loadCSS()
	if err != nil {
		return nil, err
	}
	f := template.FuncMap{
		"css":    func() template.CSS { return template.CSS(css) },
		"static": func(name string) string { return "/static/" + name + "?v=" + s.static[name].hash },
	}
	// Shared blocks: the layout and every partials*.html. Every other template
	// is a page named after its file.
	base, err := template.New("").Funcs(funcs).Funcs(f).ParseFS(assets, "templates/layout.html", "templates/partials*.html")
	if err != nil {
		return nil, err
	}
	pages := map[string]*template.Template{"fragments": base}
	files, err := fs.Glob(assets, "templates/*.html")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		name := strings.TrimSuffix(path.Base(file), ".html")
		if name == "layout" || strings.HasPrefix(name, "partials") {
			continue
		}
		t, err := template.Must(base.Clone()).ParseFS(assets, file)
		if err != nil {
			return nil, err
		}
		pages[name] = t
	}
	return pages, nil
}

var gzPool = sync.Pool{New: func() any { w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed); return w }}

// rendered is an HTML response body, plus its gzipped form when it's worth it.
type rendered struct{ raw, gz []byte }

func (s *Server) execute(page, block string, data any) (*rendered, error) {
	var buf bytes.Buffer
	if err := s.pages[page].ExecuteTemplate(&buf, block, data); err != nil {
		return nil, fmt.Errorf("render %s/%s: %w", page, block, err)
	}
	r := &rendered{raw: buf.Bytes()}
	// Below 1 KB gzip saves too little to be worth the CPU.
	if len(r.raw) >= 1024 {
		var zb bytes.Buffer
		gz := gzPool.Get().(*gzip.Writer)
		gz.Reset(&zb)
		gz.Write(r.raw)
		gz.Close()
		gzPool.Put(gz)
		r.gz = zb.Bytes()
	}
	return r, nil
}

func (s *Server) send(c *gin.Context, status int, r *rendered) {
	h := c.Writer.Header()
	h.Set("Cache-Control", "no-store")
	h.Add("Vary", "Accept-Encoding, HX-Request")
	const ctype = "text/html; charset=utf-8"
	if r.gz != nil && strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		h.Set("Content-Encoding", "gzip")
		c.Data(status, ctype, r.gz)
		return
	}
	c.Data(status, ctype, r.raw)
}

// render executes a full page ("layout").
func (s *Server) render(c *gin.Context, status int, page string, data any) {
	s.renderBlock(c, status, page, "layout", data)
}

// fragment executes one named block for an htmx swap.
func (s *Server) fragment(c *gin.Context, block string, data any) {
	s.renderBlock(c, http.StatusOK, "fragments", block, data)
}

func (s *Server) renderBlock(c *gin.Context, status int, page, block string, data any) {
	r, err := s.execute(page, block, data)
	if err != nil {
		s.fail(c, err)
		return
	}
	s.send(c, status, r)
}

// homeCache holds the home page for the current board. It only varies by whether
// the visitor has cast a ballot, and is redrawn every minute so "updated N min
// ago" stays true.
type homeCache struct {
	mu     sync.Mutex
	board  *Board
	minute int64
	pages  [2]*rendered
}

func (h *homeCache) get(b *Board, cast bool, build func() (*rendered, error)) (*rendered, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m := time.Now().Unix() / 60; h.board != b || h.minute != m {
		h.board, h.minute, h.pages = b, m, [2]*rendered{}
	}
	i := 0
	if cast {
		i = 1
	}
	if h.pages[i] == nil {
		r, err := build()
		if err != nil {
			return nil, err
		}
		h.pages[i] = r
	}
	return h.pages[i], nil
}

// --- static assets: embedded, hashed for cache busting, gzipped once at startup

type asset struct {
	body, gz []byte
	hash     string
	ctype    string
}

func loadStatic() (map[string]asset, error) {
	out := map[string]asset{}
	err := fs.WalkDir(assets, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		body, err := assets.ReadFile(p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(body)
		a := asset{body: body, hash: hex.EncodeToString(sum[:6]), ctype: mime.TypeByExtension(path.Ext(p))}
		if a.ctype == "" {
			a.ctype = "application/octet-stream"
		}
		if !strings.HasSuffix(p, ".woff2") {
			var b bytes.Buffer
			gz, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)
			gz.Write(body)
			gz.Close()
			a.gz = b.Bytes()
		}
		out[path.Base(p)] = a
		return nil
	})
	return out, err
}

func (s *Server) serveStatic(c *gin.Context) {
	a, ok := s.static[c.Param("name")]
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	h := c.Writer.Header()
	h.Set("Cache-Control", "public, max-age=31536000, immutable")
	h.Set("Vary", "Accept-Encoding")
	if a.gz != nil && strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		h.Set("Content-Encoding", "gzip")
		c.Data(http.StatusOK, a.ctype, a.gz)
		return
	}
	c.Data(http.StatusOK, a.ctype, a.body)
}
