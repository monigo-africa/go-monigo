package monigo

import (
	"context"
	"fmt"
)

// PortalTokenService manages customer portal access links for your organisation.
// Portal tokens grant an end-customer read-only access to their invoices,
// payout slips, subscriptions, and payout accounts in the Monigo hosted portal.
//
// All operations require a write-scoped API key. The organisation is derived
// automatically from the key — you never need to pass an org_id.
type PortalTokenService struct {
	client *Client
}

// Create generates a new portal link for the customer identified by
// customerExternalID.
//
// The returned PortalToken.PortalURL is the URL to share with your customer.
// Share it directly, embed it in an email, or open it inside an iframe.
//
//	token, err := client.PortalTokens.Create(ctx, monigo.CreatePortalTokenRequest{
//	    CustomerExternalID: "usr_abc123",
//	    Label:              "March 2026 invoice link",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Share this link:", token.PortalURL)
func (s *PortalTokenService) Create(ctx context.Context, req CreatePortalTokenRequest, opts ...RequestOption) (*PortalToken, error) {
	var wrapper struct {
		Token PortalToken `json:"token"`
	}
	if err := s.client.do(ctx, "POST", "/v1/portal/tokens", req, &wrapper, opts...); err != nil {
		return nil, err
	}
	return &wrapper.Token, nil
}

// List returns all active (non-revoked) portal tokens for the given customer.
// customerID may be the Monigo UUID or the customer's external_id.
func (s *PortalTokenService) List(ctx context.Context, customerID string) (*ListPortalTokensResponse, error) {
	var out ListPortalTokensResponse
	path := fmt.Sprintf("/v1/portal/tokens?customer_id=%s", customerID)
	if err := s.client.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Revoke immediately invalidates a portal token. Any customer holding the
// corresponding URL will receive a 401 on their next request.
func (s *PortalTokenService) Revoke(ctx context.Context, tokenID string) error {
	return s.client.do(ctx, "DELETE", fmt.Sprintf("/v1/portal/tokens/%s", tokenID), nil, nil)
}
