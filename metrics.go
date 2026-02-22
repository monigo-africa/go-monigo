package monigo

import (
	"context"
	"fmt"
)

// MetricService manages billing metrics — the definitions of what gets counted.
type MetricService struct {
	client *Client
}

// Create defines a new billing metric.
func (s *MetricService) Create(ctx context.Context, req CreateMetricRequest) (*Metric, error) {
	var wrapper struct {
		Metric Metric `json:"metric"`
	}
	if err := s.client.do(ctx, "POST", "/v1/metrics", req, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Metric, nil
}

// List returns all metrics for the authenticated organisation.
func (s *MetricService) List(ctx context.Context) (*ListMetricsResponse, error) {
	var out ListMetricsResponse
	if err := s.client.do(ctx, "GET", "/v1/metrics", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single metric by its UUID.
func (s *MetricService) Get(ctx context.Context, metricID string) (*Metric, error) {
	var wrapper struct {
		Metric Metric `json:"metric"`
	}
	if err := s.client.do(ctx, "GET", fmt.Sprintf("/v1/metrics/%s", metricID), nil, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Metric, nil
}

// Update modifies an existing metric's configuration.
// Note: metrics that have already been used for billing may be immutable on
// certain fields — the server will return a 400 in those cases.
func (s *MetricService) Update(ctx context.Context, metricID string, req UpdateMetricRequest) (*Metric, error) {
	var wrapper struct {
		Metric Metric `json:"metric"`
	}
	if err := s.client.do(ctx, "PUT", fmt.Sprintf("/v1/metrics/%s", metricID), req, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Metric, nil
}

// Delete permanently removes a metric record.
func (s *MetricService) Delete(ctx context.Context, metricID string) error {
	return s.client.do(ctx, "DELETE", fmt.Sprintf("/v1/metrics/%s", metricID), nil, nil)
}
