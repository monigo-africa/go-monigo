package monigo

import (
	"context"
	"fmt"
)

// PlanService manages billing plans and their associated prices.
type PlanService struct {
	client *Client
}

// Create defines a new billing plan, optionally with prices attached.
func (s *PlanService) Create(ctx context.Context, req CreatePlanRequest) (*Plan, error) {
	var wrapper struct {
		Plan Plan `json:"plan"`
	}
	if err := s.client.do(ctx, "POST", "/v1/plans", req, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Plan, nil
}

// List returns all billing plans for the authenticated organisation.
func (s *PlanService) List(ctx context.Context) (*ListPlansResponse, error) {
	var out ListPlansResponse
	if err := s.client.do(ctx, "GET", "/v1/plans", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single plan by its UUID.
func (s *PlanService) Get(ctx context.Context, planID string) (*Plan, error) {
	var wrapper struct {
		Plan Plan `json:"plan"`
	}
	if err := s.client.do(ctx, "GET", fmt.Sprintf("/v1/plans/%s", planID), nil, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Plan, nil
}

// Update modifies an existing plan's name, description, or prices.
func (s *PlanService) Update(ctx context.Context, planID string, req UpdatePlanRequest) (*Plan, error) {
	var wrapper struct {
		Plan Plan `json:"plan"`
	}
	if err := s.client.do(ctx, "PUT", fmt.Sprintf("/v1/plans/%s", planID), req, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Plan, nil
}

// Delete permanently removes a billing plan record.
func (s *PlanService) Delete(ctx context.Context, planID string) error {
	return s.client.do(ctx, "DELETE", fmt.Sprintf("/v1/plans/%s", planID), nil, nil)
}
