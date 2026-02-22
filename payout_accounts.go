package monigo

import (
	"context"
	"fmt"
)

// PayoutAccountService manages bank or mobile-money accounts for customer payouts.
// All methods require a customerID — accounts are always scoped to a customer.
type PayoutAccountService struct {
	client *Client
}

// Create adds a new payout account to a customer.
func (s *PayoutAccountService) Create(ctx context.Context, customerID string, req CreatePayoutAccountRequest) (*PayoutAccount, error) {
	var wrapper struct {
		PayoutAccount PayoutAccount `json:"payout_account"`
	}
	path := fmt.Sprintf("/v1/customers/%s/payout-accounts", customerID)
	if err := s.client.do(ctx, "POST", path, req, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.PayoutAccount, nil
}

// List returns all payout accounts for a customer.
func (s *PayoutAccountService) List(ctx context.Context, customerID string) (*ListPayoutAccountsResponse, error) {
	var out ListPayoutAccountsResponse
	path := fmt.Sprintf("/v1/customers/%s/payout-accounts", customerID)
	if err := s.client.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single payout account by its UUID.
func (s *PayoutAccountService) Get(ctx context.Context, customerID, accountID string) (*PayoutAccount, error) {
	var wrapper struct {
		PayoutAccount PayoutAccount `json:"payout_account"`
	}
	path := fmt.Sprintf("/v1/customers/%s/payout-accounts/%s", customerID, accountID)
	if err := s.client.do(ctx, "GET", path, nil, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.PayoutAccount, nil
}

// Update modifies an existing payout account.
func (s *PayoutAccountService) Update(ctx context.Context, customerID, accountID string, req UpdatePayoutAccountRequest) (*PayoutAccount, error) {
	var wrapper struct {
		PayoutAccount PayoutAccount `json:"payout_account"`
	}
	path := fmt.Sprintf("/v1/customers/%s/payout-accounts/%s", customerID, accountID)
	if err := s.client.do(ctx, "PUT", path, req, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.PayoutAccount, nil
}

// Delete permanently removes a payout account.
func (s *PayoutAccountService) Delete(ctx context.Context, customerID, accountID string) error {
	path := fmt.Sprintf("/v1/customers/%s/payout-accounts/%s", customerID, accountID)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}
