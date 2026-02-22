package monigo_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	monigo "github.com/monigo-africa/go-monigo"
)

var sampleInvoice = monigo.Invoice{
	ID:             "inv-1",
	OrgID:          "org-1",
	CustomerID:     "cust-abc",
	SubscriptionID: "sub-1",
	Status:         monigo.InvoiceStatusDraft,
	Currency:       "NGN",
	Subtotal:       "10000.00",
	Total:          "10000.00",
	PeriodStart:    time.Now().AddDate(0, -1, 0),
	PeriodEnd:      time.Now(),
	LineItems: []monigo.InvoiceLineItem{
		{
			ID:          "li-1",
			InvoiceID:   "inv-1",
			MetricID:    "metric-1",
			Description: "API Calls × 5000",
			Quantity:    "5000",
			UnitPrice:   "2.000000",
			Amount:      "10000.00",
			CreatedAt:   time.Now(),
		},
	},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

func TestInvoices_Generate(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/invoices/generate")
		assertBearerToken(t, r)

		var req monigo.GenerateInvoiceRequest
		decodeBody(t, r, &req)
		if req.SubscriptionID != "sub-1" {
			t.Errorf("subscription_id: got %q, want sub-1", req.SubscriptionID)
		}
		respondJSON(t, w, 201, map[string]any{"invoice": sampleInvoice})
	}))

	inv, err := c.Invoices.Generate(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.ID != "inv-1" {
		t.Errorf("expected inv-1, got %s", inv.ID)
	}
	if inv.Status != monigo.InvoiceStatusDraft {
		t.Errorf("expected draft, got %s", inv.Status)
	}
	if len(inv.LineItems) != 1 {
		t.Errorf("expected 1 line item, got %d", len(inv.LineItems))
	}
}

func TestInvoices_List_NoFilters(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/invoices")
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query params, got %q", r.URL.RawQuery)
		}
		respondJSON(t, w, 200, monigo.ListInvoicesResponse{
			Invoices: []monigo.Invoice{sampleInvoice},
			Count:    1,
		})
	}))

	resp, err := c.Invoices.List(context.Background(), monigo.ListInvoicesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
}

func TestInvoices_List_WithFilters(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("status") != "draft" {
			t.Errorf("status: got %q, want draft", q.Get("status"))
		}
		if q.Get("customer_id") != "cust-abc" {
			t.Errorf("customer_id: got %q, want cust-abc", q.Get("customer_id"))
		}
		respondJSON(t, w, 200, monigo.ListInvoicesResponse{
			Invoices: []monigo.Invoice{sampleInvoice},
			Count:    1,
		})
	}))

	_, err := c.Invoices.List(context.Background(), monigo.ListInvoicesParams{
		Status:     "draft",
		CustomerID: "cust-abc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvoices_Get(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/invoices/inv-1")
		respondJSON(t, w, 200, map[string]any{"invoice": sampleInvoice})
	}))

	inv, err := c.Invoices.Get(context.Background(), "inv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Total != "10000.00" {
		t.Errorf("expected total 10000.00, got %s", inv.Total)
	}
}

func TestInvoices_Get_NotFound(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "invoice not found")
	}))
	_, err := c.Invoices.Get(context.Background(), "missing")
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}

func TestInvoices_Finalize(t *testing.T) {
	finalized := sampleInvoice
	finalized.Status = monigo.InvoiceStatusFinalized
	now := time.Now()
	finalized.FinalizedAt = &now

	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/invoices/inv-1/finalize")
		respondJSON(t, w, 200, map[string]any{"invoice": finalized})
	}))

	inv, err := c.Invoices.Finalize(context.Background(), "inv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != monigo.InvoiceStatusFinalized {
		t.Errorf("expected finalized, got %s", inv.Status)
	}
	if inv.FinalizedAt == nil {
		t.Error("expected FinalizedAt to be set")
	}
}

func TestInvoices_Void(t *testing.T) {
	voided := sampleInvoice
	voided.Status = monigo.InvoiceStatusVoid

	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/invoices/inv-1/void")
		respondJSON(t, w, 200, map[string]any{"invoice": voided})
	}))

	inv, err := c.Invoices.Void(context.Background(), "inv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != monigo.InvoiceStatusVoid {
		t.Errorf("expected void, got %s", inv.Status)
	}
}
