package monigo_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	monigo "github.com/monigo-africa/go-monigo"
)

var sampleWallet = monigo.CustomerWallet{
	ID:              "wal-1",
	CustomerID:      "cust-abc",
	OrgID:           "org-1",
	Currency:        "NGN",
	Balance:         "500.000000",
	ReservedBalance: "0.000000",
	CreatedAt:       time.Now(),
	UpdatedAt:       time.Now(),
}

var sampleVirtualAccount = monigo.VirtualAccount{
	ID:            "va-1",
	CustomerID:    "cust-abc",
	WalletID:      "wal-1",
	OrgID:         "org-1",
	Provider:      "paystack",
	AccountNumber: "1234567890",
	AccountName:   "Acme Corp",
	BankName:      "Access Bank",
	BankCode:      "044",
	Currency:      "NGN",
	ProviderRef:   "ref_123",
	IsActive:      true,
	CreatedAt:     time.Now(),
	UpdatedAt:     time.Now(),
}

var sampleLedgerEntry = monigo.LedgerEntry{
	ID:             "le-1",
	OrgID:          "org-1",
	TransactionID:  "txn-1",
	AccountType:    "customer_wallet",
	AccountID:      "wal-1",
	Direction:      "credit",
	Amount:         "100.000000",
	Currency:       "NGN",
	BalanceBefore:  "400.000000",
	BalanceAfter:   "500.000000",
	Description:    "Top-up",
	EntryType:      "deposit",
	ReferenceType:  "manual_credit",
	ReferenceID:    "ref-1",
	IdempotencyKey: "idem-1",
	CreatedAt:      time.Now(),
}

func TestWallets_GetOrCreate(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/wallets")
		assertBearerToken(t, r)

		var req monigo.GetOrCreateWalletRequest
		decodeBody(t, r, &req)
		if req.CustomerID != "cust-abc" {
			t.Errorf("customer_id: got %q, want cust-abc", req.CustomerID)
		}
		respondJSON(t, w, 200, map[string]any{"wallet": sampleWallet})
	}))

	wallet, err := c.Wallets.GetOrCreate(context.Background(), monigo.GetOrCreateWalletRequest{
		CustomerID: "cust-abc",
		Currency:   "NGN",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wallet.ID != "wal-1" {
		t.Errorf("expected ID wal-1, got %s", wallet.ID)
	}
}

func TestWallets_List(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/wallets")
		respondJSON(t, w, 200, monigo.ListWalletsResponse{
			Wallets: []monigo.CustomerWallet{sampleWallet},
			Count:   1,
		})
	}))

	resp, err := c.Wallets.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
}

func TestWallets_ListByCustomerID(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/wallets")
		if r.URL.Query().Get("customer_id") != "cust-abc" {
			t.Errorf("expected customer_id=cust-abc, got %s", r.URL.Query().Get("customer_id"))
		}
		respondJSON(t, w, 200, monigo.ListWalletsResponse{
			Wallets: []monigo.CustomerWallet{sampleWallet},
			Count:   1,
		})
	}))

	resp, err := c.Wallets.List(context.Background(), monigo.ListWalletsParams{CustomerID: "cust-abc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
}

func TestWallets_ListByCustomer(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/customers/cust-abc/wallets")
		respondJSON(t, w, 200, monigo.ListWalletsResponse{
			Wallets: []monigo.CustomerWallet{sampleWallet},
			Count:   1,
		})
	}))

	resp, err := c.Wallets.ListByCustomer(context.Background(), "cust-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Wallets) != 1 {
		t.Errorf("expected 1 wallet, got %d", len(resp.Wallets))
	}
}

func TestWallets_Get(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/wallets/wal-1")
		respondJSON(t, w, 200, monigo.WalletWithVirtualAccountsResponse{
			Wallet:          sampleWallet,
			VirtualAccounts: []monigo.VirtualAccount{sampleVirtualAccount},
		})
	}))

	resp, err := c.Wallets.Get(context.Background(), "wal-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Wallet.ID != "wal-1" {
		t.Errorf("expected wallet ID wal-1, got %s", resp.Wallet.ID)
	}
	if len(resp.VirtualAccounts) != 1 {
		t.Errorf("expected 1 virtual account, got %d", len(resp.VirtualAccounts))
	}
}

func TestWallets_Get_NotFound(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 404, "wallet not found")
	}))
	_, err := c.Wallets.Get(context.Background(), "missing")
	if !monigo.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}

func TestWallets_Credit(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/wallets/wal-1/credit")

		var req monigo.CreditWalletRequest
		decodeBody(t, r, &req)
		if req.Amount != "100.000000" {
			t.Errorf("amount: got %q, want 100.000000", req.Amount)
		}
		respondJSON(t, w, 200, monigo.WalletOperationResponse{
			Wallet:        sampleWallet,
			LedgerEntries: []monigo.LedgerEntry{sampleLedgerEntry},
		})
	}))

	resp, err := c.Wallets.Credit(context.Background(), "wal-1", monigo.CreditWalletRequest{
		Amount:         "100.000000",
		Currency:       "NGN",
		Description:    "Top-up",
		EntryType:      monigo.WalletEntryTypeDeposit,
		ReferenceType:  "manual_credit",
		ReferenceID:    "ref-1",
		IdempotencyKey: "idem-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Wallet.ID != "wal-1" {
		t.Errorf("expected wallet ID wal-1, got %s", resp.Wallet.ID)
	}
	if len(resp.LedgerEntries) != 1 {
		t.Errorf("expected 1 ledger entry, got %d", len(resp.LedgerEntries))
	}
}

func TestWallets_Debit(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/wallets/wal-1/debit")
		respondJSON(t, w, 200, monigo.WalletOperationResponse{
			Wallet:        sampleWallet,
			LedgerEntries: []monigo.LedgerEntry{sampleLedgerEntry},
		})
	}))

	resp, err := c.Wallets.Debit(context.Background(), "wal-1", monigo.DebitWalletRequest{
		Amount:         "50.000000",
		Currency:       "NGN",
		Description:    "Usage charge",
		EntryType:      monigo.WalletEntryTypeUsage,
		ReferenceType:  "usage_event",
		ReferenceID:    "evt-1",
		IdempotencyKey: "idem-2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Wallet.ID != "wal-1" {
		t.Errorf("expected wallet ID wal-1, got %s", resp.Wallet.ID)
	}
}

func TestWallets_Debit_InsufficientBalance(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondError(t, w, 402, "insufficient wallet balance")
	}))
	_, err := c.Wallets.Debit(context.Background(), "wal-1", monigo.DebitWalletRequest{
		Amount:         "999999.000000",
		Currency:       "NGN",
		Description:    "Too much",
		EntryType:      monigo.WalletEntryTypeUsage,
		ReferenceType:  "usage_event",
		ReferenceID:    "evt-2",
		IdempotencyKey: "idem-3",
	})
	if !monigo.IsQuotaExceeded(err) {
		t.Errorf("expected IsQuotaExceeded=true (402); err=%v", err)
	}
}

func TestWallets_ListTransactions(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/wallets/wal-1/transactions")
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}
		respondJSON(t, w, 200, monigo.ListTransactionsResponse{
			Transactions: []monigo.LedgerEntry{sampleLedgerEntry},
			Total:        1,
			Limit:        10,
			Offset:       0,
		})
	}))

	resp, err := c.Wallets.ListTransactions(context.Background(), "wal-1", monigo.ListTransactionsParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
}

func TestWallets_CreateVirtualAccount(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		assertPath(t, r, "/v1/wallets/wal-1/virtual-accounts")

		var req monigo.CreateVirtualAccountRequest
		decodeBody(t, r, &req)
		if req.Provider != "paystack" {
			t.Errorf("provider: got %q, want paystack", req.Provider)
		}
		respondJSON(t, w, 201, map[string]any{"virtual_account": sampleVirtualAccount})
	}))

	va, err := c.Wallets.CreateVirtualAccount(context.Background(), "wal-1", monigo.CreateVirtualAccountRequest{
		Provider: monigo.VirtualAccountProviderPaystack,
		Currency: "NGN",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if va.ID != "va-1" {
		t.Errorf("expected ID va-1, got %s", va.ID)
	}
}

func TestWallets_ListVirtualAccounts(t *testing.T) {
	c := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		assertPath(t, r, "/v1/wallets/wal-1/virtual-accounts")
		respondJSON(t, w, 200, monigo.ListVirtualAccountsResponse{
			VirtualAccounts: []monigo.VirtualAccount{sampleVirtualAccount},
			Count:           1,
		})
	}))

	resp, err := c.Wallets.ListVirtualAccounts(context.Background(), "wal-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
	if resp.VirtualAccounts[0].Provider != "paystack" {
		t.Errorf("expected provider paystack, got %s", resp.VirtualAccounts[0].Provider)
	}
}
