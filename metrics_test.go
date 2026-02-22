package monigo_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	monigo "github.com/monigo-africa/go-monigo"
)

var sampleMetric = monigo.Metric{
	ID:          "metric-1",
	OrgID:       "org-1",
	Name:        "API Calls",
	EventName:   "api_call",
	Aggregation: monigo.AggregationCount,
	CreatedAt:   time.Now(),
	UpdatedAt:   time.Now(),
}

func TestMetrics_Create(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/metrics")

		var req monigo.CreateMetricRequest
		decodeBody(t, r, &req)
		if req.Aggregation != monigo.AggregationCount {
			t.Errorf("aggregation: got %q, want %q", req.Aggregation, monigo.AggregationCount)
		}
		respondJSON(t, w, 201, map[string]any{"metric": sampleMetric})
	}))

	m, err := c.Metrics.Create(context.Background(), monigo.CreateMetricRequest{
		Name:        "API Calls",
		EventName:   "api_call",
		Aggregation: monigo.AggregationCount,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != "metric-1" {
		t.Errorf("expected ID metric-1, got %s", m.ID)
	}
}

func TestMetrics_List(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/metrics")
		respondJSON(t, w, 200, monigo.ListMetricsResponse{
			Metrics: []monigo.Metric{sampleMetric},
			Count:   1,
		})
	}))

	resp, err := c.Metrics.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
}

func TestMetrics_Get(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/metrics/metric-1")
		respondJSON(t, w, 200, map[string]any{"metric": sampleMetric})
	}))

	m, err := c.Metrics.Get(context.Background(), "metric-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.EventName != "api_call" {
		t.Errorf("expected api_call, got %s", m.EventName)
	}
}

func TestMetrics_Get_NotFound(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "metric not found")
	}))
	_, err := c.Metrics.Get(context.Background(), "x")
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}

func TestMetrics_Update(t *testing.T) {
	updated := sampleMetric
	updated.Description = "Counts API calls"

	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "PUT")
		assertPath(t, r, "/v1/metrics/metric-1")
		respondJSON(t, w, 200, map[string]any{"metric": updated})
	}))

	m, err := c.Metrics.Update(context.Background(), "metric-1", monigo.UpdateMetricRequest{
		Description: "Counts API calls",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Description != "Counts API calls" {
		t.Errorf("expected description, got %s", m.Description)
	}
}

func TestMetrics_Delete(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "DELETE")
		assertPath(t, r, "/v1/metrics/metric-1")
		respondJSON(t, w, 200, map[string]string{"message": "Metric deleted successfully"})
	}))

	if err := c.Metrics.Delete(context.Background(), "metric-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
