package store

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptrace"
	"sync"
	"syscall"
	"time"
)

// The libsql driver sends every statement as an HTTP POST through
// http.DefaultClient and has no option for its own client. Turso closes idle
// keep-alive connections, and a statement that goes out on one just as it closes
// fails with EOF. net/http never retries a POST, so without this the first page
// after a few idle minutes was a 500.
var useStaleRetry sync.Once

func installStaleRetry() {
	useStaleRetry.Do(func() {
		base := http.DefaultTransport.(*http.Transport).Clone()
		// Drop idle connections well before the server does, so most never go stale.
		base.IdleConnTimeout = 20 * time.Second
		// Only the driver uses DefaultClient; the TMDB client has its own.
		http.DefaultClient.Transport = &staleRetry{base: base}
	})
}

// staleRetry resends a request once, on a fresh connection, when it failed on
// a reused connection that the server had already closed. The server never read
// such a request, so sending it again can't apply a statement twice.
type staleRetry struct {
	base interface {
		http.RoundTripper
		CloseIdleConnections()
	}
}

func (t *staleRetry) RoundTrip(req *http.Request) (*http.Response, error) {
	var reused bool
	trace := &httptrace.ClientTrace{GotConn: func(i httptrace.GotConnInfo) { reused = i.Reused }}
	resp, err := t.base.RoundTrip(req.WithContext(httptrace.WithClientTrace(req.Context(), trace)))
	if err == nil || !reused || !staleConn(err) || req.GetBody == nil {
		return resp, err
	}
	body, gerr := req.GetBody()
	if gerr != nil {
		return nil, err
	}
	// Other pooled connections went idle at the same time; don't retry on one.
	t.base.CloseIdleConnections()
	retry := req.Clone(req.Context())
	retry.Body = body
	return t.base.RoundTrip(retry)
}

func staleConn(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE)
}
