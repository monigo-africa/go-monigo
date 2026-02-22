package monigo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultBaseURL = "https://api.monigo.co"

// Client is the Monigo API client. Create one with New() and use its
// resource services to interact with the API.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	// Events handles usage event ingestion and event replay.
	Events *EventService
	// Customers manages your end-customers.
	Customers *CustomerService
	// Metrics manages billing metrics (what gets counted).
	Metrics *MetricService
	// Plans manages billing plans and their prices.
	Plans *PlanService
	// Subscriptions links customers to plans.
	Subscriptions *SubscriptionService
	// PayoutAccounts manages bank/mobile-money accounts for customer payouts.
	PayoutAccounts *PayoutAccountService
	// Invoices manages invoice generation, finalization, and voiding.
	Invoices *InvoiceService
	// Usage queries usage rollups per customer/metric.
	Usage *UsageService
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL (https://api.monigo.co).
// Useful for self-hosted deployments or pointing at a local dev server.
func WithBaseURL(u string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(u, "/")
	}
}

// WithHTTPClient replaces the default http.Client with a custom one.
// Use this to set timeouts, custom transports, or proxies.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// New creates a new Monigo API client authenticated with apiKey.
// Pass functional options to override defaults.
//
//	client := monigo.New(os.Getenv("MONIGO_API_KEY"))
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{},
	}
	for _, o := range opts {
		o(c)
	}
	c.Events = &EventService{client: c}
	c.Customers = &CustomerService{client: c}
	c.Metrics = &MetricService{client: c}
	c.Plans = &PlanService{client: c}
	c.Subscriptions = &SubscriptionService{client: c}
	c.PayoutAccounts = &PayoutAccountService{client: c}
	c.Invoices = &InvoiceService{client: c}
	c.Usage = &UsageService{client: c}
	return c
}

// do executes an HTTP request against the Monigo API.
//
// method is the HTTP method (GET, POST, PUT, PATCH, DELETE).
// path must start with "/", e.g. "/v1/customers".
// body is marshalled to JSON and sent as the request body (pass nil for no body).
// out is decoded from the JSON response body (pass nil to discard response body).
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("monigo: marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("monigo: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("monigo: execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("monigo: read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		// Try to decode structured error; fall back to raw body.
		if jsonErr := json.Unmarshal(respBody, apiErr); jsonErr != nil {
			apiErr.Message = string(respBody)
		}
		return apiErr
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("monigo: decode response: %w", err)
		}
	}
	return nil
}
