package monigo

import (
	"context"
	"fmt"
	"net/url"
)

// InvoiceService manages invoice generation, retrieval, finalization, and voiding.
type InvoiceService struct {
	client *Client
}

// Generate creates a new draft invoice for the given subscription based on
// current period usage. The invoice starts in "draft" status.
func (s *InvoiceService) Generate(ctx context.Context, subscriptionID string, opts ...RequestOption) (*Invoice, error) {
	var wrapper struct {
		Invoice Invoice `json:"invoice"`
	}
	body := GenerateInvoiceRequest{SubscriptionID: subscriptionID}
	if err := s.client.do(ctx, "POST", "/v1/invoices/generate", body, &wrapper, opts...); err != nil {
		return nil, err
	}
	return &wrapper.Invoice, nil
}

// List returns invoices, optionally filtered by status or customer.
func (s *InvoiceService) List(ctx context.Context, params ListInvoicesParams) (*ListInvoicesResponse, error) {
	q := url.Values{}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.CustomerID != "" {
		q.Set("customer_id", params.CustomerID)
	}

	path := "/v1/invoices"
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}

	var out ListInvoicesResponse
	if err := s.client.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single invoice by its UUID, including line items.
func (s *InvoiceService) Get(ctx context.Context, invoiceID string) (*Invoice, error) {
	var wrapper struct {
		Invoice Invoice `json:"invoice"`
	}
	if err := s.client.do(ctx, "GET", fmt.Sprintf("/v1/invoices/%s", invoiceID), nil, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Invoice, nil
}

// Finalize transitions a draft invoice to "finalized", making it ready for payment.
// A finalized invoice cannot be edited.
func (s *InvoiceService) Finalize(ctx context.Context, invoiceID string, opts ...RequestOption) (*Invoice, error) {
	var wrapper struct {
		Invoice Invoice `json:"invoice"`
	}
	if err := s.client.do(ctx, "POST", fmt.Sprintf("/v1/invoices/%s/finalize", invoiceID), nil, &wrapper, opts...); err != nil {
		return nil, err
	}
	return &wrapper.Invoice, nil
}

// Void marks an invoice as void, making it no longer payable.
func (s *InvoiceService) Void(ctx context.Context, invoiceID string, opts ...RequestOption) (*Invoice, error) {
	var wrapper struct {
		Invoice Invoice `json:"invoice"`
	}
	if err := s.client.do(ctx, "POST", fmt.Sprintf("/v1/invoices/%s/void", invoiceID), nil, &wrapper, opts...); err != nil {
		return nil, err
	}
	return &wrapper.Invoice, nil
}
