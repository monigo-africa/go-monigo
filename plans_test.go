package monigo_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	monigo "github.com/monigo-africa/go-monigo"
)

var samplePlan = monigo.Plan{
	ID:            "plan-1",
	OrgID:         "org-1",
	Name:          "API Pro",
	Currency:      "NGN",
	PlanType:      monigo.PlanTypeCollection,
	BillingPeriod: monigo.BillingPeriodMonthly,
	CreatedAt:     time.Now(),
	UpdatedAt:     time.Now(),
}

func TestPlans_Create(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/plans")

		var req monigo.CreatePlanRequest
		decodeBody(t, r, &req)
		if req.Name != "API Pro" {
			t.Errorf("name: got %q, want API Pro", req.Name)
		}
		if req.BillingPeriod != monigo.BillingPeriodMonthly {
			t.Errorf("billing_period: got %q, want monthly", req.BillingPeriod)
		}
		respondJSON(t, w, 201, map[string]any{"plan": samplePlan})
	}))

	plan, err := c.Plans.Create(context.Background(), monigo.CreatePlanRequest{
		Name:          "API Pro",
		Currency:      "NGN",
		PlanType:      monigo.PlanTypeCollection,
		BillingPeriod: monigo.BillingPeriodMonthly,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.ID != "plan-1" {
		t.Errorf("expected plan-1, got %s", plan.ID)
	}
}

func TestPlans_Create_WithPrices(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req monigo.CreatePlanRequest
		decodeBody(t, r, &req)
		if len(req.Prices) != 1 {
			t.Errorf("expected 1 price, got %d", len(req.Prices))
		}
		if req.Prices[0].Model != monigo.PricingModelFlat {
			t.Errorf("model: got %q, want %q", req.Prices[0].Model, monigo.PricingModelFlat)
		}
		respondJSON(t, w, 201, map[string]any{"plan": samplePlan})
	}))

	_, err := c.Plans.Create(context.Background(), monigo.CreatePlanRequest{
		Name: "API Pro",
		Prices: []monigo.CreatePriceRequest{
			{MetricID: "m-1", Model: monigo.PricingModelFlat, UnitPrice: "2.000000"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlans_Create_WithTieredPrices(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req monigo.CreatePlanRequest
		decodeBody(t, r, &req)
		if len(req.Prices) != 1 {
			t.Errorf("expected 1 price")
		}
		var tiers []monigo.PriceTier
		if err := json.Unmarshal(req.Prices[0].Tiers, &tiers); err != nil {
			t.Fatalf("unmarshal tiers: %v", err)
		}
		if len(tiers) != 2 {
			t.Errorf("expected 2 tiers, got %d", len(tiers))
		}
		respondJSON(t, w, 201, map[string]any{"plan": samplePlan})
	}))

	limit := int64(1000)
	tiersJSON, _ := json.Marshal([]monigo.PriceTier{
		{UpTo: &limit, UnitAmount: "1.000000"},
		{UpTo: nil, UnitAmount: "0.500000"},
	})
	_, err := c.Plans.Create(context.Background(), monigo.CreatePlanRequest{
		Name: "Tiered Plan",
		Prices: []monigo.CreatePriceRequest{
			{
				MetricID: "m-1",
				Model:    monigo.PricingModelTiered,
				Tiers:    tiersJSON,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlans_List(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/plans")
		respondJSON(t, w, 200, monigo.ListPlansResponse{Plans: []monigo.Plan{samplePlan}, Count: 1})
	}))

	resp, err := c.Plans.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
}

func TestPlans_Get(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertPath(t, r, "/v1/plans/plan-1")
		respondJSON(t, w, 200, map[string]any{"plan": samplePlan})
	}))

	plan, err := c.Plans.Get(context.Background(), "plan-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Name != "API Pro" {
		t.Errorf("expected API Pro, got %s", plan.Name)
	}
}

func TestPlans_Get_NotFound(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "plan not found")
	}))
	_, err := c.Plans.Get(context.Background(), "x")
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}

func TestPlans_Update(t *testing.T) {
	updated := samplePlan
	updated.Name = "API Pro Plus"

	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "PUT")
		assertPath(t, r, "/v1/plans/plan-1")
		respondJSON(t, w, 200, map[string]any{"plan": updated})
	}))

	plan, err := c.Plans.Update(context.Background(), "plan-1", monigo.UpdatePlanRequest{Name: "API Pro Plus"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Name != "API Pro Plus" {
		t.Errorf("expected API Pro Plus, got %s", plan.Name)
	}
}

func TestPlans_Delete(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "DELETE")
		assertPath(t, r, "/v1/plans/plan-1")
		respondJSON(t, w, 200, map[string]string{"message": "Plan deleted successfully"})
	}))

	if err := c.Plans.Delete(context.Background(), "plan-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
