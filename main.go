// Command filmstop serves the Filmstop site.
//
//	PORT            listen port (default 8080)
//	DATABASE_PATH   SQLite file (default filmstop.db)
//	TMDB_API_KEY    TMDB v3 key or v4 read token; search is off without it
//
// "filmstop -healthcheck" exits 0 if the server on PORT answers /healthz. The
// container image has no shell or curl, so Docker's HEALTHCHECK uses this.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Binamra7/filmstop/internal/store"
	"github.com/Binamra7/filmstop/internal/tmdb"
	"github.com/Binamra7/filmstop/internal/web"
	"github.com/gin-gonic/gin"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	healthcheck := flag.Bool("healthcheck", false, "probe the running server and exit")
	flag.Parse()
	if *healthcheck {
		c := http.Client{Timeout: 2 * time.Second}
		res, err := c.Get("http://127.0.0.1:" + env("PORT", "8080") + "/healthz")
		if err != nil || res.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	gin.SetMode(gin.ReleaseMode)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := store.Open(store.Source(env("DATABASE_PATH", "filmstop.db")))
	if err != nil {
		slog.Error("open database", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	tm := tmdb.New(os.Getenv("TMDB_API_KEY"))
	if !tm.Enabled() {
		slog.Warn("TMDB_API_KEY is not set; film search is off")
	}
	srv, err := web.New(ctx, st, tm)
	if err != nil {
		slog.Error("start", "err", err)
		os.Exit(1)
	}

	hs := &http.Server{
		Addr:              ":" + env("PORT", "8080"),
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	go func() {
		slog.Info("listening", "addr", hs.Addr)
		if err := hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("serve", "err", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hs.Shutdown(shutdown)
}
