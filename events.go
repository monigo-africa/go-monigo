package monigo

import (
	"context"
	"fmt"
	"time"
)

// EventService handles usage event ingestion and event replay.
type EventService struct {
	client *Client
}

// Ingest sends one or more usage events to the Monigo ingestion pipeline.
// Events are processed asynchronously; the response confirms receipt.
//
// Each event must have a unique IdempotencyKey — resending the same key is
// safe and will be de-duplicated server-side.
//
// Requires an API key with the "ingest" scope.
func (s *EventService) Ingest(ctx context.Context, req IngestRequest, opts ...RequestOption) (*IngestResponse, error) {
	var wrapper struct {
		Ingested   []string `json:"ingested"`
		Duplicates []string `json:"duplicates"`
	}
	if err := s.client.do(ctx, "POST", "/v1/ingest", req, &wrapper, opts...); err != nil {
		return nil, err
	}
	return &IngestResponse{
		Ingested:   wrapper.Ingested,
		Duplicates: wrapper.Duplicates,
	}, nil
}

// StartReplay initiates an asynchronous replay of all raw events in the
// given time window through the current processing pipeline.
//
// eventName is optional; pass nil to replay all event types in the window.
//
// Returns a job record immediately — poll GetReplay to track progress.
func (s *EventService) StartReplay(ctx context.Context, from, to time.Time, eventName *string, opts ...RequestOption) (*EventReplayJob, error) {
	body := map[string]any{
		"from": from.Format(time.RFC3339),
		"to":   to.Format(time.RFC3339),
	}
	if eventName != nil {
		body["event_name"] = *eventName
	}

	var wrapper struct {
		Job EventReplayJob `json:"job"`
	}
	if err := s.client.do(ctx, "POST", "/v1/events/replay", body, &wrapper, opts...); err != nil {
		return nil, err
	}
	return &wrapper.Job, nil
}

// GetReplay fetches the current status of an event replay job.
func (s *EventService) GetReplay(ctx context.Context, jobID string) (*EventReplayJob, error) {
	var wrapper struct {
		Job EventReplayJob `json:"job"`
	}
	if err := s.client.do(ctx, "GET", fmt.Sprintf("/v1/events/replay/%s", jobID), nil, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Job, nil
}
