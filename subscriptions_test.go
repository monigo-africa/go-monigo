package monigo_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	monigo "github.com/monigo-africa/go-monigo"
)

var sampleSubscription = monigo.Subscription{
	ID:                 "sub-1",
	OrgID:              "org-1",
	CustomerID:         "cust-abc",
	PlanID:             "plan-1",
	Status:             monigo.SubscriptionStatusActive,
	CurrentPeriodStart: time.Now(),
	CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
	CreatedAt:          time.Now(),
	UpdatedAt:          time.Now(),
}

func TestSubscriptions_Create(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/subscriptions")

		var req monigo.CreateSubscriptionRequest
		decodeBody(t, r, &req)
		if req.CustomerID != "cust-abc" {
			t.Errorf("customer_id: got %q, want cust-abc", req.CustomerID)
		}
		if req.PlanID != "plan-1" {
			t.Errorf("plan_id: got %q, want plan-1", req.PlanID)
		}
		respondJSON(t, w, 201, map[string]any{"subscription": sampleSubscription})
	}))

	sub, err := c.Subscriptions.Create(context.Background(), monigo.CreateSubscriptionRequest{
		CustomerID: "cust-abc",
		PlanID:     "plan-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Errorf("expected sub-1, got %s", sub.ID)
	}
	if sub.Status != monigo.SubscriptionStatusActive {
		t.Errorf("expected active, got %s", sub.Status)
	}
}

func TestSubscriptions_Create_Conflict(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 409, "customer already has an active subscription")
	}))

	_, err := c.Subscriptions.Create(context.Background(), monigo.CreateSubscriptionRequest{
		CustomerID: "cust-abc",
		PlanID:     "plan-1",
	})
	if !monigo.IsConflict(err) {
		t.Errorf("expected IsConflict=true; err=%v", err)
	}
}

func TestSubscriptions_List_NoParams(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/subscriptions")
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query params, got %q", r.URL.RawQuery)
		}
		respondJSON(t, w, 200, monigo.ListSubscriptionsResponse{
			Subscriptions: []monigo.Subscription{sampleSubscription},
			Count:         1,
		})
	}))

	resp, err := c.Subscriptions.List(context.Background(), monigo.ListSubscriptionsParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
}

func TestSubscriptions_List_WithParams(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		if r.URL.Query().Get("customer_id") != "cust-abc" {
			t.Errorf("expected customer_id=cust-abc, got %q", r.URL.Query().Get("customer_id"))
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("expected status=active, got %q", r.URL.Query().Get("status"))
		}
		respondJSON(t, w, 200, monigo.ListSubscriptionsResponse{
			Subscriptions: []monigo.Subscription{sampleSubscription},
			Count:         1,
		})
	}))

	_, err := c.Subscriptions.List(context.Background(), monigo.ListSubscriptionsParams{
		CustomerID: "cust-abc",
		Status:     "active",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubscriptions_Get(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/subscriptions/sub-1")
		respondJSON(t, w, 200, map[string]any{"subscription": sampleSubscription})
	}))

	sub, err := c.Subscriptions.Get(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.PlanID != "plan-1" {
		t.Errorf("expected plan-1, got %s", sub.PlanID)
	}
}

func TestSubscriptions_UpdateStatus(t *testing.T) {
	paused := sampleSubscription
	paused.Status = monigo.SubscriptionStatusPaused

	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "PATCH")
		assertPath(t, r, "/v1/subscriptions/sub-1")

		var body map[string]string
		decodeBody(t, r, &body)
		if body["status"] != "paused" {
			t.Errorf("status: got %q, want paused", body["status"])
		}
		respondJSON(t, w, 200, map[string]any{"subscription": paused})
	}))

	sub, err := c.Subscriptions.UpdateStatus(context.Background(), "sub-1", monigo.SubscriptionStatusPaused)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Status != monigo.SubscriptionStatusPaused {
		t.Errorf("expected paused, got %s", sub.Status)
	}
}

func TestSubscriptions_Delete(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "DELETE")
		assertPath(t, r, "/v1/subscriptions/sub-1")
		respondJSON(t, w, 200, map[string]string{"message": "Subscription cancelled successfully"})
	}))

	if err := c.Subscriptions.Delete(context.Background(), "sub-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubscriptions_Get_NotFound(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "subscription not found")
	}))
	_, err := c.Subscriptions.Get(context.Background(), "missing")
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}
