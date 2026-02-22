package monigo_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	monigo "github.com/monigo-africa/go-monigo"
)

func TestEvents_Ingest_Success(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/ingest")
		assertBearerToken(t, r)

		var body monigo.IngestRequest
		decodeBody(t, r, &body)
		if len(body.Events) != 2 {
			t.Errorf("expected 2 events, got %d", len(body.Events))
		}
		respondJSON(t, w, 202, map[string]any{
			"ingested":   []string{"key-1", "key-2"},
			"duplicates": []string{},
		})
	}))

	now := time.Now()
	resp, err := c.Events.Ingest(context.Background(), monigo.IngestRequest{
		Events: []monigo.IngestEvent{
			{EventName: "api_call", CustomerID: "cust-1", IdempotencyKey: "key-1", Timestamp: now, Properties: map[string]any{}},
			{EventName: "api_call", CustomerID: "cust-1", IdempotencyKey: "key-2", Timestamp: now, Properties: map[string]any{}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Ingested) != 2 {
		t.Errorf("expected 2 ingested, got %d", len(resp.Ingested))
	}
	if len(resp.Duplicates) != 0 {
		t.Errorf("expected 0 duplicates, got %d", len(resp.Duplicates))
	}
}

func TestEvents_Ingest_WithDuplicates(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(t, w, 202, map[string]any{
			"ingested":   []string{"key-1"},
			"duplicates": []string{"key-2"},
		})
	}))

	now := time.Now()
	resp, err := c.Events.Ingest(context.Background(), monigo.IngestRequest{
		Events: []monigo.IngestEvent{
			{EventName: "api_call", CustomerID: "c1", IdempotencyKey: "key-1", Timestamp: now},
			{EventName: "api_call", CustomerID: "c1", IdempotencyKey: "key-2", Timestamp: now},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Ingested) != 1 || resp.Ingested[0] != "key-1" {
		t.Errorf("unexpected ingested: %v", resp.Ingested)
	}
	if len(resp.Duplicates) != 1 || resp.Duplicates[0] != "key-2" {
		t.Errorf("unexpected duplicates: %v", resp.Duplicates)
	}
}

func TestEvents_Ingest_QuotaExceeded(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 402, "quota exceeded")
	}))

	_, err := c.Events.Ingest(context.Background(), monigo.IngestRequest{
		Events: []monigo.IngestEvent{{EventName: "api_call", CustomerID: "c1", IdempotencyKey: "k", Timestamp: time.Now()}},
	})
	if !monigo.IsQuotaExceeded(err) {
		t.Errorf("expected IsQuotaExceeded=true, got false; err=%v", err)
	}
}

func TestEvents_StartReplay(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/events/replay")

		var body map[string]any
		decodeBody(t, r, &body)
		if body["from"] == nil || body["to"] == nil {
			t.Error("expected from and to in body")
		}
		if _, ok := body["event_name"]; ok {
			t.Error("event_name should not be set when nil")
		}

		respondJSON(t, w, 202, map[string]any{
			"job": monigo.EventReplayJob{
				ID:     "job-1",
				Status: "pending",
			},
		})
	}))

	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()
	job, err := c.Events.StartReplay(context.Background(), from, to, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID != "job-1" {
		t.Errorf("expected job ID job-1, got %s", job.ID)
	}
	if job.Status != "pending" {
		t.Errorf("expected status pending, got %s", job.Status)
	}
}

func TestEvents_StartReplay_WithEventName(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decodeBody(t, r, &body)
		if body["event_name"] != "api_call" {
			t.Errorf("expected event_name=api_call, got %v", body["event_name"])
		}
		respondJSON(t, w, 202, map[string]any{"job": monigo.EventReplayJob{ID: "job-2", Status: "pending"}})
	}))

	name := "api_call"
	_, err := c.Events.StartReplay(context.Background(), time.Now().Add(-time.Hour), time.Now(), &name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvents_GetReplay(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/events/replay/job-99")
		respondJSON(t, w, 200, map[string]any{
			"job": monigo.EventReplayJob{
				ID:             "job-99",
				Status:         "completed",
				EventsTotal:    100,
				EventsReplayed: 100,
			},
		})
	}))

	job, err := c.Events.GetReplay(context.Background(), "job-99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != "completed" {
		t.Errorf("expected status completed, got %s", job.Status)
	}
	if job.EventsReplayed != 100 {
		t.Errorf("expected 100 replayed, got %d", job.EventsReplayed)
	}
}

func TestEvents_GetReplay_NotFound(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "job not found")
	}))
	_, err := c.Events.GetReplay(context.Background(), "missing")
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}
