package monigo

import (
	"context"
	"fmt"
)

// CustomerService manages the end-customers in your Monigo organisation.
type CustomerService struct {
	client *Client
}

// Create registers a new customer.
func (s *CustomerService) Create(ctx context.Context, req CreateCustomerRequest) (*Customer, error) {
	var wrapper struct {
		Customer Customer `json:"customer"`
	}
	if err := s.client.do(ctx, "POST", "/v1/customers", req, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Customer, nil
}

// List returns all customers belonging to the authenticated organisation.
func (s *CustomerService) List(ctx context.Context) (*ListCustomersResponse, error) {
	var out ListCustomersResponse
	if err := s.client.do(ctx, "GET", "/v1/customers", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single customer by their Monigo UUID.
func (s *CustomerService) Get(ctx context.Context, customerID string) (*Customer, error) {
	var wrapper struct {
		Customer Customer `json:"customer"`
	}
	if err := s.client.do(ctx, "GET", fmt.Sprintf("/v1/customers/%s", customerID), nil, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Customer, nil
}

// Update modifies an existing customer's name, email, or metadata.
// Only non-zero fields in req are sent; pass zero values to leave fields unchanged.
func (s *CustomerService) Update(ctx context.Context, customerID string, req UpdateCustomerRequest) (*Customer, error) {
	var wrapper struct {
		Customer Customer `json:"customer"`
	}
	if err := s.client.do(ctx, "PUT", fmt.Sprintf("/v1/customers/%s", customerID), req, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Customer, nil
}

// Delete permanently removes a customer record.
func (s *CustomerService) Delete(ctx context.Context, customerID string) error {
	return s.client.do(ctx, "DELETE", fmt.Sprintf("/v1/customers/%s", customerID), nil, nil)
}
