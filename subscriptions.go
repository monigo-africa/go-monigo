package monigo

import (
	"context"
	"fmt"
	"net/url"
)

// SubscriptionService links customers to billing plans.
type SubscriptionService struct {
	client *Client
}

// Create subscribes a customer to a plan. Returns a 409 Conflict error
// (use IsConflict) if the customer already has an active subscription.
func (s *SubscriptionService) Create(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error) {
	var wrapper struct {
		Subscription Subscription `json:"subscription"`
	}
	if err := s.client.do(ctx, "POST", "/v1/subscriptions", req, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Subscription, nil
}

// List returns subscriptions, optionally filtered by customer, plan, or status.
func (s *SubscriptionService) List(ctx context.Context, params ListSubscriptionsParams) (*ListSubscriptionsResponse, error) {
	q := url.Values{}
	if params.CustomerID != "" {
		q.Set("customer_id", params.CustomerID)
	}
	if params.PlanID != "" {
		q.Set("plan_id", params.PlanID)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}

	path := "/v1/subscriptions"
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}

	var out ListSubscriptionsResponse
	if err := s.client.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single subscription by its UUID.
func (s *SubscriptionService) Get(ctx context.Context, subscriptionID string) (*Subscription, error) {
	var wrapper struct {
		Subscription Subscription `json:"subscription"`
	}
	if err := s.client.do(ctx, "GET", fmt.Sprintf("/v1/subscriptions/%s", subscriptionID), nil, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Subscription, nil
}

// UpdateStatus changes the status of a subscription.
// Use the SubscriptionStatusXxx constants: active, paused, canceled.
func (s *SubscriptionService) UpdateStatus(ctx context.Context, subscriptionID, status string) (*Subscription, error) {
	body := map[string]string{"status": status}
	var wrapper struct {
		Subscription Subscription `json:"subscription"`
	}
	if err := s.client.do(ctx, "PATCH", fmt.Sprintf("/v1/subscriptions/%s", subscriptionID), body, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Subscription, nil
}

// Delete cancels and removes a subscription record.
func (s *SubscriptionService) Delete(ctx context.Context, subscriptionID string) error {
	return s.client.do(ctx, "DELETE", fmt.Sprintf("/v1/subscriptions/%s", subscriptionID), nil, nil)
}
