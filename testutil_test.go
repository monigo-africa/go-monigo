package monigo_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	monigo "github.com/monigo-africa/go-monigo"
)

// mockServer spins up an in-process HTTP server backed by handler, creates a
// Client pointed at it, and registers cleanup to shut down the server when the
// test finishes.
func mockServer(t *testing.T, handler http.Handler) *monigo.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return monigo.New("test_key_abc", monigo.WithBaseURL(srv.URL))
}

// respondJSON writes status and v (encoded as JSON) to w.
func respondJSON(t *testing.T, w http.ResponseWriter, status int, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("respondJSON encode: %v", err)
	}
}

// respondError writes a standard Monigo error envelope.
func respondError(t *testing.T, w http.ResponseWriter, status int, message string) {
	t.Helper()
	respondJSON(t, w, status, map[string]string{"error": message})
}

// decodeBody reads and decodes the JSON request body from r into v.
func decodeBody(t *testing.T, r *http.Request, v any) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		t.Fatalf("decodeBody: %v", err)
	}
}

// assertMethod fails the test if r.Method != want.
func assertMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.Method != want {
		t.Errorf("method: got %q, want %q", r.Method, want)
	}
}

// assertPath fails the test if r.URL.Path != want.
func assertPath(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.URL.Path != want {
		t.Errorf("path: got %q, want %q", r.URL.Path, want)
	}
}

// assertBearerToken fails the test if the Authorization header != "Bearer test_key_abc".
func assertBearerToken(t *testing.T, r *http.Request) {
	t.Helper()
	got := r.Header.Get("Authorization")
	if got != "Bearer test_key_abc" {
		t.Errorf("authorization header: got %q, want %q", got, "Bearer test_key_abc")
	}
}
