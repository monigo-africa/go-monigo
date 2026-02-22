package monigo_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	monigo "github.com/monigo-africa/go-monigo"
)

func TestUsage_Query_NoParams(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/usage")
		// No query params when nothing is set
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query params, got %q", r.URL.RawQuery)
		}
		respondJSON(t, w, 200, monigo.UsageQueryResult{
			Rollups: []monigo.UsageRollup{
				{
					ID:          "rollup-1",
					CustomerID:  "cust-abc",
					MetricID:    "metric-1",
					Aggregation: monigo.AggregationCount,
					Value:       5000,
					EventCount:  5000,
					PeriodStart: time.Now().AddDate(0, -1, 0),
					PeriodEnd:   time.Now(),
				},
			},
			Count: 1,
		})
	}))

	result, err := c.Usage.Query(context.Background(), monigo.UsageParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 1 {
		t.Errorf("expected count 1, got %d", result.Count)
	}
	if result.Rollups[0].Value != 5000 {
		t.Errorf("expected value 5000, got %f", result.Rollups[0].Value)
	}
}

func TestUsage_Query_WithCustomerAndMetric(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("customer_id") != "cust-abc" {
			t.Errorf("customer_id: got %q, want cust-abc", q.Get("customer_id"))
		}
		if q.Get("metric_id") != "metric-1" {
			t.Errorf("metric_id: got %q, want metric-1", q.Get("metric_id"))
		}
		respondJSON(t, w, 200, monigo.UsageQueryResult{Count: 0, Rollups: []monigo.UsageRollup{}})
	}))

	_, err := c.Usage.Query(context.Background(), monigo.UsageParams{
		CustomerID: "cust-abc",
		MetricID:   "metric-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUsage_Query_WithTimeRange(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("from") == "" {
			t.Error("expected from param to be set")
		}
		if q.Get("to") == "" {
			t.Error("expected to param to be set")
		}
		respondJSON(t, w, 200, monigo.UsageQueryResult{Count: 0, Rollups: []monigo.UsageRollup{}})
	}))

	_, err := c.Usage.Query(context.Background(), monigo.UsageParams{
		From: &from,
		To:   &to,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUsage_Query_EmptyResult(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(t, w, 200, monigo.UsageQueryResult{Count: 0, Rollups: []monigo.UsageRollup{}})
	}))

	result, err := c.Usage.Query(context.Background(), monigo.UsageParams{CustomerID: "unknown"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 0 {
		t.Errorf("expected count 0, got %d", result.Count)
	}
	if len(result.Rollups) != 0 {
		t.Errorf("expected empty rollups, got %d", len(result.Rollups))
	}
}

func TestUsage_Query_Unauthorized(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 401, "unauthorized")
	}))
	_, err := c.Usage.Query(context.Background(), monigo.UsageParams{})
	if !monigo.IsUnauthorized(err) {
		t.Errorf("expected IsUnauthorized=true; err=%v", err)
	}
}
