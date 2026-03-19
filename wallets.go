package monigo

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// WalletService manages customer wallets, wallet operations, and virtual accounts.
type WalletService struct {
	client *Client
}

// GetOrCreate retrieves an existing wallet or creates a new one for the given
// customer and currency combination.
func (s *WalletService) GetOrCreate(ctx context.Context, req GetOrCreateWalletRequest, opts ...RequestOption) (*CustomerWallet, error) {
	var wrapper struct {
		Wallet CustomerWallet `json:"wallet"`
	}
	if err := s.client.do(ctx, "POST", "/v1/wallets", req, &wrapper, opts...); err != nil {
		return nil, err
	}
	return &wrapper.Wallet, nil
}

// List returns all wallets for the authenticated organisation.
// Pass an optional ListWalletsParams to filter by customer.
func (s *WalletService) List(ctx context.Context, params ...ListWalletsParams) (*ListWalletsResponse, error) {
	path := "/v1/wallets"
	if len(params) > 0 && params[0].CustomerID != "" {
		q := url.Values{}
		q.Set("customer_id", params[0].CustomerID)
		path = path + "?" + q.Encode()
	}

	var out ListWalletsResponse
	if err := s.client.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListByCustomer returns all wallets belonging to a specific customer.
func (s *WalletService) ListByCustomer(ctx context.Context, customerID string) (*ListWalletsResponse, error) {
	var out ListWalletsResponse
	if err := s.client.do(ctx, "GET", fmt.Sprintf("/v1/customers/%s/wallets", customerID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single wallet by its UUID, including its virtual accounts.
func (s *WalletService) Get(ctx context.Context, walletID string) (*WalletWithVirtualAccountsResponse, error) {
	var out WalletWithVirtualAccountsResponse
	if err := s.client.do(ctx, "GET", fmt.Sprintf("/v1/wallets/%s", walletID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Credit adds funds to a wallet and returns the updated wallet with ledger entries.
func (s *WalletService) Credit(ctx context.Context, walletID string, req CreditWalletRequest, opts ...RequestOption) (*WalletOperationResponse, error) {
	var out WalletOperationResponse
	if err := s.client.do(ctx, "POST", fmt.Sprintf("/v1/wallets/%s/credit", walletID), req, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// Debit removes funds from a wallet and returns the updated wallet with ledger entries.
// Returns a 402 error if the wallet has insufficient balance.
func (s *WalletService) Debit(ctx context.Context, walletID string, req DebitWalletRequest, opts ...RequestOption) (*WalletOperationResponse, error) {
	var out WalletOperationResponse
	if err := s.client.do(ctx, "POST", fmt.Sprintf("/v1/wallets/%s/debit", walletID), req, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTransactions returns paginated ledger entries for a wallet.
func (s *WalletService) ListTransactions(ctx context.Context, walletID string, params ListTransactionsParams) (*ListTransactionsResponse, error) {
	q := url.Values{}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}

	path := fmt.Sprintf("/v1/wallets/%s/transactions", walletID)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}

	var out ListTransactionsResponse
	if err := s.client.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateVirtualAccount provisions a dedicated virtual bank account that
// automatically funds the wallet on deposit.
func (s *WalletService) CreateVirtualAccount(ctx context.Context, walletID string, req CreateVirtualAccountRequest, opts ...RequestOption) (*VirtualAccount, error) {
	var wrapper struct {
		VirtualAccount VirtualAccount `json:"virtual_account"`
	}
	if err := s.client.do(ctx, "POST", fmt.Sprintf("/v1/wallets/%s/virtual-accounts", walletID), req, &wrapper, opts...); err != nil {
		return nil, err
	}
	return &wrapper.VirtualAccount, nil
}

// ListVirtualAccounts returns all virtual accounts linked to a wallet.
func (s *WalletService) ListVirtualAccounts(ctx context.Context, walletID string) (*ListVirtualAccountsResponse, error) {
	var out ListVirtualAccountsResponse
	if err := s.client.do(ctx, "GET", fmt.Sprintf("/v1/wallets/%s/virtual-accounts", walletID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
