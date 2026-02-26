package monigo_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	monigo "github.com/monigo-africa/go-monigo"
)

var sampleToken = monigo.PortalToken{
	ID:         "tok-1",
	OrgID:      "org-1",
	CustomerID: "cust-abc",
	Token:      "aabbccdd1122334455667788aabbccdd1122334455667788aabbccdd11223344",
	Label:      "Invoice link",
	PortalURL:  "https://app.monigo.co/portal/aabbccdd1122334455667788aabbccdd1122334455667788aabbccdd11223344",
	CreatedAt:  time.Now(),
	UpdatedAt:  time.Now(),
}

func TestPortalTokens_Create(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/portal/tokens")
		assertBearerToken(t, r)

		var req monigo.CreatePortalTokenRequest
		decodeBody(t, r, &req)
		if req.CustomerExternalID != "usr_abc123" {
			t.Errorf("customer_external_id: got %q, want usr_abc123", req.CustomerExternalID)
		}
		if req.Label != "Invoice link" {
			t.Errorf("label: got %q, want Invoice link", req.Label)
		}
		respondJSON(t, w, 201, map[string]any{
			"token":      sampleToken,
			"portal_url": sampleToken.PortalURL,
		})
	}))

	tok, err := c.PortalTokens.Create(context.Background(), monigo.CreatePortalTokenRequest{
		CustomerExternalID: "usr_abc123",
		Label:              "Invoice link",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.ID != "tok-1" {
		t.Errorf("expected tok-1, got %s", tok.ID)
	}
	if tok.PortalURL == "" {
		t.Error("expected non-empty portal_url")
	}
}

func TestPortalTokens_Create_WithExpiry(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		var req monigo.CreatePortalTokenRequest
		decodeBody(t, r, &req)
		if req.ExpiresAt == "" {
			t.Error("expected expires_at to be set")
		}
		respondJSON(t, w, 201, map[string]any{"token": sampleToken, "portal_url": sampleToken.PortalURL})
	}))

	_, err := c.PortalTokens.Create(context.Background(), monigo.CreatePortalTokenRequest{
		CustomerExternalID: "usr_abc123",
		ExpiresAt:          "2027-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPortalTokens_List(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertBearerToken(t, r)
		if r.URL.Path != "/v1/portal/tokens" {
			t.Errorf("path: got %q, want /v1/portal/tokens", r.URL.Path)
		}
		if q := r.URL.Query().Get("customer_id"); q != "cust-abc" {
			t.Errorf("customer_id: got %q, want cust-abc", q)
		}
		respondJSON(t, w, 200, monigo.ListPortalTokensResponse{
			Tokens: []monigo.PortalToken{sampleToken},
			Count:  1,
		})
	}))

	resp, err := c.PortalTokens.List(context.Background(), "cust-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
	if resp.Tokens[0].ID != "tok-1" {
		t.Errorf("unexpected token ID: %s", resp.Tokens[0].ID)
	}
}

func TestPortalTokens_Revoke(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "DELETE")
		assertPath(t, r, "/v1/portal/tokens/tok-1")
		assertBearerToken(t, r)
		respondJSON(t, w, 200, map[string]string{"message": "Portal token revoked successfully"})
	}))

	if err := c.PortalTokens.Revoke(context.Background(), "tok-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPortalTokens_Revoke_NotFound(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "portal token not found")
	}))

	err := c.PortalTokens.Revoke(context.Background(), "missing")
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}

func TestPortalTokens_Create_CustomerNotFound(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "customer not found")
	}))

	_, err := c.PortalTokens.Create(context.Background(), monigo.CreatePortalTokenRequest{
		CustomerExternalID: "nonexistent",
	})
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}
