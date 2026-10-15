package web

import (
	"context"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Binamra7/filmstop/internal/store"
	"github.com/Binamra7/filmstop/internal/tmdb"
	"github.com/gin-gonic/gin"
)

// BenchmarkHome renders the home page with a full 30-film board, gzipped.
func BenchmarkHome(b *testing.B) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	st, _ := store.Open(filepath.Join(b.TempDir(), "b.db"))
	defer st.Close()
	var ms []store.Movie
	for i := int64(1); i <= 60; i++ {
		ms = append(ms, store.Movie{ID: i, Title: fmt.Sprintf("Film %d", i), Year: "1999", PosterPath: "/p.jpg", Overview: "An overview."})
	}
	st.UpsertMovies(ctx, ms)
	for v := range 300 {
		var bl store.Ballot
		for i := range bl {
			bl[i] = int64((v*7+i*11)%60 + 1)
		}
		st.Cast(ctx, fmt.Sprintf("v%031d", v), bl)
	}
	srv, err := New(ctx, st, tmdb.New(""))
	if err != nil {
		b.Fatal(err)
	}
	h := srv.Handler()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Accept-Encoding", "gzip")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != 200 {
				b.Fatal(w.Code)
			}
		}
	})
}
