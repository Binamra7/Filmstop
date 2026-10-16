package store

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptrace"
	"testing"
)

// fakeBase plays a transport whose first call fails on a connection it reports
// as reused or fresh.
type fakeBase struct {
	firstErr    error
	firstReused bool
	calls       int
	closed      int
	bodies      []string
}

func (f *fakeBase) RoundTrip(req *http.Request) (*http.Response, error) {
	f.calls++
	b, _ := io.ReadAll(req.Body)
	f.bodies = append(f.bodies, string(b))
	if f.calls == 1 {
		if tr := httptrace.ContextClientTrace(req.Context()); tr != nil && tr.GotConn != nil {
			tr.GotConn(httptrace.GotConnInfo{Reused: f.firstReused})
		}
		if f.firstErr != nil {
			return nil, f.firstErr
		}
	}
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}

func (f *fakeBase) CloseIdleConnections() { f.closed++ }

func post(t *testing.T, rt http.RoundTripper) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, "https://db.example/v2/pipeline", bytes.NewReader([]byte("stmt")))
	if err != nil {
		t.Fatal(err)
	}
	return rt.RoundTrip(req)
}

func TestStaleRetry(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		reused    bool
		wantCalls int
		wantErr   bool
	}{
		{"stale reused connection is retried", io.EOF, true, 2, false},
		{"reset reused connection is retried", &netErr{io.ErrUnexpectedEOF}, true, 2, false},
		{"fresh connection is not retried", io.EOF, false, 1, true},
		{"other errors are not retried", errors.New("dns failure"), true, 1, true},
		{"success is passed through", nil, true, 1, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeBase{firstErr: c.err, firstReused: c.reused}
			_, err := post(t, &staleRetry{base: f})
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, want error %v", err, c.wantErr)
			}
			if f.calls != c.wantCalls {
				t.Fatalf("calls = %d, want %d", f.calls, c.wantCalls)
			}
			if c.wantCalls == 2 {
				if f.bodies[1] != "stmt" {
					t.Fatalf("retry body = %q, want the original", f.bodies[1])
				}
				if f.closed != 1 {
					t.Fatalf("idle connections closed %d times before retry, want 1", f.closed)
				}
			}
		})
	}
}

// netErr wraps an error the way net/http reports a broken connection.
type netErr struct{ err error }

func (e *netErr) Error() string { return "read: " + e.err.Error() }
func (e *netErr) Unwrap() error { return e.err }
