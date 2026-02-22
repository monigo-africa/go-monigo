package monigo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	monigo "github.com/monigo-africa/go-monigo"
)

func TestNew_Defaults(t *testing.T) {
	c := monigo.New("sk_test")
	if c == nil {
		t.Fatal("New returned nil")
	}
	// All services must be initialised.
	if c.Events == nil {
		t.Error("Events service is nil")
	}
	if c.Customers == nil {
		t.Error("Customers service is nil")
	}
	if c.Metrics == nil {
		t.Error("Metrics service is nil")
	}
	if c.Plans == nil {
		t.Error("Plans service is nil")
	}
	if c.Subscriptions == nil {
		t.Error("Subscriptions service is nil")
	}
	if c.PayoutAccounts == nil {
		t.Error("PayoutAccounts service is nil")
	}
	if c.Invoices == nil {
		t.Error("Invoices service is nil")
	}
	if c.Usage == nil {
		t.Error("Usage service is nil")
	}
}

func TestWithBaseURL(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		respondJSON(t, w, 200, map[string]any{"customers": []any{}, "count": 0})
	}))
	defer srv.Close()

	c := monigo.New("sk_test", monigo.WithBaseURL(srv.URL))
	_, err := c.Customers.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("server was not called — WithBaseURL not applied")
	}
}

func TestWithHTTPClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(t, w, 200, map[string]any{"customers": []any{}, "count": 0})
	}))
	defer srv.Close()

	custom := &http.Client{}
	c := monigo.New("sk_test", monigo.WithBaseURL(srv.URL), monigo.WithHTTPClient(custom))
	_, err := c.Customers.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDo_SetsAuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertBearerToken(t, r)
		respondJSON(t, w, 200, map[string]any{"customers": []any{}, "count": 0})
	}))
	defer srv.Close()

	c := monigo.New("test_key_abc", monigo.WithBaseURL(srv.URL))
	_, _ = c.Customers.List(context.Background())
}

func TestDo_Returns404AsAPIError(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "not found")
	}))
	_, err := c.Customers.Get(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false; err=%v", err)
	}
}

func TestDo_Returns401AsAPIError(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 401, "unauthorized")
	}))
	_, err := c.Customers.Get(context.Background(), "x")
	if !monigo.IsUnauthorized(err) {
		t.Errorf("expected IsUnauthorized=true, got false; err=%v", err)
	}
}

func TestDo_Returns429AsRateLimited(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 429, "too many requests")
	}))
	_, err := c.Customers.Get(context.Background(), "x")
	if !monigo.IsRateLimited(err) {
		t.Errorf("expected IsRateLimited=true, got false; err=%v", err)
	}
}

func TestDo_Returns500AsAPIError(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 500, "internal server error")
	}))
	_, err := c.Customers.Get(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
