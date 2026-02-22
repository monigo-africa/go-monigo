package monigo_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	monigo "github.com/monigo-africa/go-monigo"
)

var sampleAccount = monigo.PayoutAccount{
	ID:            "acct-1",
	CustomerID:    "cust-abc",
	OrgID:         "org-1",
	AccountName:   "John Driver",
	BankName:      "First Bank Nigeria",
	BankCode:      "011",
	AccountNumber: "3001234567",
	PayoutMethod:  monigo.PayoutMethodBankTransfer,
	Currency:      "NGN",
	IsDefault:     true,
	CreatedAt:     time.Now(),
	UpdatedAt:     time.Now(),
}

func TestPayoutAccounts_Create(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/customers/cust-abc/payout-accounts")
		assertBearerToken(t, r)

		var req monigo.CreatePayoutAccountRequest
		decodeBody(t, r, &req)
		if req.PayoutMethod != monigo.PayoutMethodBankTransfer {
			t.Errorf("payout_method: got %q, want bank_transfer", req.PayoutMethod)
		}
		respondJSON(t, w, 201, map[string]any{"payout_account": sampleAccount})
	}))

	acct, err := c.PayoutAccounts.Create(context.Background(), "cust-abc", monigo.CreatePayoutAccountRequest{
		AccountName:   "John Driver",
		PayoutMethod:  monigo.PayoutMethodBankTransfer,
		BankName:      "First Bank Nigeria",
		BankCode:      "011",
		AccountNumber: "3001234567",
		Currency:      "NGN",
		IsDefault:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acct.ID != "acct-1" {
		t.Errorf("expected acct-1, got %s", acct.ID)
	}
}

func TestPayoutAccounts_List(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/customers/cust-abc/payout-accounts")
		respondJSON(t, w, 200, monigo.ListPayoutAccountsResponse{
			PayoutAccounts: []monigo.PayoutAccount{sampleAccount},
			Count:          1,
		})
	}))

	resp, err := c.PayoutAccounts.List(context.Background(), "cust-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
	if resp.PayoutAccounts[0].AccountName != "John Driver" {
		t.Errorf("unexpected account name: %s", resp.PayoutAccounts[0].AccountName)
	}
}

func TestPayoutAccounts_Get(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/customers/cust-abc/payout-accounts/acct-1")
		respondJSON(t, w, 200, map[string]any{"payout_account": sampleAccount})
	}))

	acct, err := c.PayoutAccounts.Get(context.Background(), "cust-abc", "acct-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acct.BankCode != "011" {
		t.Errorf("expected bank code 011, got %s", acct.BankCode)
	}
}

func TestPayoutAccounts_Get_NotFound(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "payout account not found")
	}))
	_, err := c.PayoutAccounts.Get(context.Background(), "cust-abc", "missing")
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}

func TestPayoutAccounts_Update(t *testing.T) {
	updated := sampleAccount
	updated.AccountName = "Jane Driver"

	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "PUT")
		assertPath(t, r, "/v1/customers/cust-abc/payout-accounts/acct-1")

		var req monigo.UpdatePayoutAccountRequest
		decodeBody(t, r, &req)
		if req.AccountName != "Jane Driver" {
			t.Errorf("account_name: got %q, want Jane Driver", req.AccountName)
		}
		respondJSON(t, w, 200, map[string]any{"payout_account": updated})
	}))

	acct, err := c.PayoutAccounts.Update(context.Background(), "cust-abc", "acct-1",
		monigo.UpdatePayoutAccountRequest{AccountName: "Jane Driver"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acct.AccountName != "Jane Driver" {
		t.Errorf("expected Jane Driver, got %s", acct.AccountName)
	}
}

func TestPayoutAccounts_Delete(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "DELETE")
		assertPath(t, r, "/v1/customers/cust-abc/payout-accounts/acct-1")
		respondJSON(t, w, 200, map[string]string{"message": "Payout account deleted successfully"})
	}))

	if err := c.PayoutAccounts.Delete(context.Background(), "cust-abc", "acct-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
