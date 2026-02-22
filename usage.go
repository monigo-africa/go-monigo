package monigo

import (
	"context"
	"net/url"
	"time"
)

// UsageService queries usage rollups aggregated from ingested events.
type UsageService struct {
	client *Client
}

// Query returns per-customer, per-metric usage rollups for the organisation.
// All fields in UsageParams are optional; omit them to get the full current billing period.
func (s *UsageService) Query(ctx context.Context, params UsageParams) (*UsageQueryResult, error) {
	q := url.Values{}
	if params.CustomerID != "" {
		q.Set("customer_id", params.CustomerID)
	}
	if params.MetricID != "" {
		q.Set("metric_id", params.MetricID)
	}
	if params.From != nil {
		q.Set("from", params.From.UTC().Format(time.RFC3339))
	}
	if params.To != nil {
		q.Set("to", params.To.UTC().Format(time.RFC3339))
	}

	path := "/v1/usage"
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}

	var out UsageQueryResult
	if err := s.client.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
