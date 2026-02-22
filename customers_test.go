package monigo_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	monigo "github.com/monigo-africa/go-monigo"
)

var sampleCustomer = monigo.Customer{
	ID:         "cust-abc",
	OrgID:      "org-1",
	ExternalID: "ext-1",
	Name:       "Acme Corp",
	Email:      "acme@example.com",
	CreatedAt:  time.Now(),
	UpdatedAt:  time.Now(),
}

func TestCustomers_Create(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/customers")
		assertBearerToken(t, r)

		var req monigo.CreateCustomerRequest
		decodeBody(t, r, &req)
		if req.ExternalID != "ext-1" {
			t.Errorf("external_id: got %q, want ext-1", req.ExternalID)
		}
		respondJSON(t, w, 201, map[string]any{"customer": sampleCustomer})
	}))

	cust, err := c.Customers.Create(context.Background(), monigo.CreateCustomerRequest{
		ExternalID: "ext-1",
		Name:       "Acme Corp",
		Email:      "acme@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cust.ID != "cust-abc" {
		t.Errorf("expected ID cust-abc, got %s", cust.ID)
	}
}

func TestCustomers_List(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/customers")
		respondJSON(t, w, 200, monigo.ListCustomersResponse{
			Customers: []monigo.Customer{sampleCustomer},
			Count:     1,
		})
	}))

	resp, err := c.Customers.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
	if len(resp.Customers) != 1 {
		t.Errorf("expected 1 customer, got %d", len(resp.Customers))
	}
}

func TestCustomers_Get(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/customers/cust-abc")
		respondJSON(t, w, 200, map[string]any{"customer": sampleCustomer})
	}))

	cust, err := c.Customers.Get(context.Background(), "cust-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cust.Name != "Acme Corp" {
		t.Errorf("expected Acme Corp, got %s", cust.Name)
	}
}

func TestCustomers_Get_NotFound(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "customer not found")
	}))
	_, err := c.Customers.Get(context.Background(), "missing")
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}

func TestCustomers_Update(t *testing.T) {
	updated := sampleCustomer
	updated.Name = "Acme Updated"

	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "PUT")
		assertPath(t, r, "/v1/customers/cust-abc")

		var req monigo.UpdateCustomerRequest
		decodeBody(t, r, &req)
		if req.Name != "Acme Updated" {
			t.Errorf("name: got %q, want Acme Updated", req.Name)
		}
		respondJSON(t, w, 200, map[string]any{"customer": updated})
	}))

	cust, err := c.Customers.Update(context.Background(), "cust-abc", monigo.UpdateCustomerRequest{Name: "Acme Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cust.Name != "Acme Updated" {
		t.Errorf("expected Acme Updated, got %s", cust.Name)
	}
}

func TestCustomers_Delete(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "DELETE")
		assertPath(t, r, "/v1/customers/cust-abc")
		respondJSON(t, w, 200, map[string]string{"message": "Customer deleted successfully"})
	}))

	if err := c.Customers.Delete(context.Background(), "cust-abc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCustomers_Delete_NotFound(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "customer not found")
	}))
	err := c.Customers.Delete(context.Background(), "missing")
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}
