package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient spins up an httptest.Server with the given handler and returns
// a Client wired to it. The server is closed automatically when the test ends.
func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		token:      "test-token",
		baseURL:    srv.URL,
		httpClient: srv.Client(),
	}
}

// respond returns a simple handler that asserts the request method and path,
// checks the auth header, and writes the given status + body.
func respond(t *testing.T, wantMethod, wantPath string, status int, body string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		if r.Method != wantMethod {
			t.Errorf("method: got %q, want %q", r.Method, wantMethod)
		}
		if r.URL.Path != wantPath {
			t.Errorf("path: got %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.Header.Get("X-Seal-Token"); got != "test-token" {
			t.Errorf("X-Seal-Token: got %q, want %q", got, "test-token")
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}
}

// respondWithBodyCheck is like respond but also calls bodyFn with the decoded
// request body so callers can assert request payload shape.
func respondWithBodyCheck(
	t *testing.T,
	wantMethod, wantPath string,
	status int,
	respBody string,
	bodyFn func(t *testing.T, body map[string]any),
) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		if r.Method != wantMethod {
			t.Errorf("method: got %q, want %q", r.Method, wantMethod)
		}
		if r.URL.Path != wantPath {
			t.Errorf("path: got %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.Header.Get("X-Seal-Token"); got != "test-token" {
			t.Errorf("X-Seal-Token: got %q, want %q", got, "test-token")
		}
		if bodyFn != nil {
			raw, _ := io.ReadAll(r.Body)
			var m map[string]any
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Errorf("request body not valid JSON: %v\nbody: %s", err, raw)
			} else {
				bodyFn(t, m)
			}
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, respBody)
	}
}

// ─── Generic error-path tests ─────────────────────────────────────────────────

func TestDo_503RateLimitError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(503)
	})
	_, err := c.GetSubscription(1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("expected 503 message, got: %v", err)
	}
}

func TestDo_4xxAPIError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		_, _ = io.WriteString(w, `{"error":"not found"}`)
	})
	_, err := c.GetSubscription(999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 in error, got: %v", err)
	}
}

func TestDo_AuthHeader(t *testing.T) {
	var capturedToken string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedToken = r.Header.Get("X-Seal-Token")
		w.WriteHeader(200)
		_, _ = io.WriteString(w, `{}`)
	})
	_, _ = c.GetSubscription(1)
	if capturedToken != "test-token" {
		t.Errorf("X-Seal-Token: got %q, want %q", capturedToken, "test-token")
	}
}

func TestDo_TrailingSlashOnPath(t *testing.T) {
	var capturedPath string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		_, _ = io.WriteString(w, `{}`)
	})
	_, _ = c.ListSubscriptions(ListSubscriptionsParams{})
	if !strings.HasSuffix(capturedPath, "/") {
		t.Errorf("expected trailing slash on path, got: %q", capturedPath)
	}
}

func TestDo_SuccessReturnsRawBytes(t *testing.T) {
	const want = `{"subscriptions":[]}`
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = io.WriteString(w, want)
	})
	got, err := c.ListSubscriptions(ListSubscriptionsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("body: got %q, want %q", got, want)
	}
}
