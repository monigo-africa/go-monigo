package monigo

import (
	"encoding/json"
	"time"
)

// ---------------------------------------------------------------------------
// Aggregation constants
// ---------------------------------------------------------------------------

const (
	AggregationCount   = "count"
	AggregationSum     = "sum"
	AggregationMax     = "max"
	AggregationMin     = "minimum"
	AggregationAverage = "average"
	AggregationUnique  = "unique"
)

// ---------------------------------------------------------------------------
// Pricing model constants
// ---------------------------------------------------------------------------

const (
	// PricingModelFlat charges a fixed unit_price per unit, regardless of volume.
	PricingModelFlat = "flat_unit"
	// PricingModelPerUnit is an alias for PricingModelFlat.
	PricingModelPerUnit = "per_unit"
	// PricingModelTiered applies graduated rates: each unit is charged at the
	// rate of the tier it falls into. Requires a []PriceTier in Tiers.
	PricingModelTiered = "tiered"
	// PricingModelPackage charges per bundle of N units. Partial bundles are
	// rounded up. Requires a PackageConfig in Tiers.
	PricingModelPackage = "package"
	// PricingModelOverage includes a free quota (IncludedUnits) covered by a
	// flat BasePrice, then charges OveragePrice per unit beyond the quota.
	// Requires an OverageConfig in Tiers.
	PricingModelOverage = "overage"
)

// ---------------------------------------------------------------------------
// Plan constants
// ---------------------------------------------------------------------------

const (
	PlanTypeCollection = "collection"
	PlanTypePayout     = "payout"

	BillingPeriodDaily     = "daily"
	BillingPeriodWeekly    = "weekly"
	BillingPeriodMonthly   = "monthly"
	BillingPeriodQuarterly = "quarterly"
	BillingPeriodAnnually  = "annually"
)

// ---------------------------------------------------------------------------
// Subscription status constants
// ---------------------------------------------------------------------------

const (
	SubscriptionStatusActive   = "active"
	SubscriptionStatusPaused   = "paused"
	SubscriptionStatusCanceled = "canceled"
)

// ---------------------------------------------------------------------------
// Invoice status constants
// ---------------------------------------------------------------------------

const (
	InvoiceStatusDraft     = "draft"
	InvoiceStatusFinalized = "finalized"
	InvoiceStatusPaid      = "paid"
	InvoiceStatusVoid      = "void"
)

// ---------------------------------------------------------------------------
// Payout method constants
// ---------------------------------------------------------------------------

const (
	PayoutMethodBankTransfer  = "bank_transfer"
	PayoutMethodMobileMoney   = "mobile_money"
)

// ---------------------------------------------------------------------------
// Ingest types
// ---------------------------------------------------------------------------

// IngestEvent represents a single usage event to be ingested.
type IngestEvent struct {
	// EventName is the name of the event (e.g. "api_call", "storage.write").
	EventName string `json:"event_name"`
	// CustomerID is the Monigo customer UUID this event belongs to.
	CustomerID string `json:"customer_id"`
	// IdempotencyKey is a unique identifier for this event. Re-sending the
	// same key is safe — the server will de-duplicate automatically.
	IdempotencyKey string `json:"idempotency_key"`
	// Timestamp is when the event occurred. Backdated events are allowed
	// within the configured replay window.
	Timestamp time.Time `json:"timestamp"`
	// Properties is an arbitrary map of key-value pairs attached to the event.
	// Use this for dimensions like endpoint, region, tier, etc.
	Properties map[string]any `json:"properties"`
}

// IngestRequest is the body sent to POST /v1/ingest.
type IngestRequest struct {
	Events []IngestEvent `json:"events"`
}

// IngestResponse is returned by POST /v1/ingest.
type IngestResponse struct {
	// Ingested contains the IdempotencyKeys of events that were successfully ingested.
	Ingested []string `json:"ingested"`
	// Duplicates contains the IdempotencyKeys of events that were skipped
	// because they were already ingested.
	Duplicates []string `json:"duplicates"`
}

// ---------------------------------------------------------------------------
// Customer types
// ---------------------------------------------------------------------------

// Customer represents an end-customer record inside your Monigo organisation.
type Customer struct {
	ID         string          `json:"id"`
	OrgID      string          `json:"org_id"`
	ExternalID string          `json:"external_id"`
	Name       string          `json:"name"`
	Email      string          `json:"email"`
	// Phone is the customer's phone number in E.164 format (e.g. +2348012345678).
	Phone      string          `json:"phone"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// CreateCustomerRequest is the body for POST /v1/customers.
type CreateCustomerRequest struct {
	// ExternalID is the ID used for this customer in your own system.
	ExternalID string `json:"external_id"`
	// Name is the customer's display name.
	Name string `json:"name"`
	// Email is the customer's email address.
	Email string `json:"email,omitempty"`
	// Phone is the customer's phone number in E.164 format (e.g. +2348012345678). Optional.
	Phone string `json:"phone,omitempty"`
	// Metadata is an optional JSON blob of arbitrary data.
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// UpdateCustomerRequest is the body for PUT /v1/customers/{id}.
// Only fields with non-zero values are updated.
type UpdateCustomerRequest struct {
	Name     string          `json:"name,omitempty"`
	Email    string          `json:"email,omitempty"`
	// Phone is the customer's phone number in E.164 format (e.g. +2348012345678). Optional.
	Phone    string          `json:"phone,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// ListCustomersResponse is returned by GET /v1/customers.
type ListCustomersResponse struct {
	Customers []Customer `json:"customers"`
	Count     int        `json:"count"`
}

// ---------------------------------------------------------------------------
// Metric types
// ---------------------------------------------------------------------------

// Metric defines what usage is counted and how.
type Metric struct {
	ID                  string    `json:"id"`
	OrgID               string    `json:"org_id"`
	Name                string    `json:"name"`
	EventName           string    `json:"event_name"`
	Aggregation         string    `json:"aggregation"`
	AggregationProperty string    `json:"aggregation_property,omitempty"`
	Description         string    `json:"description,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// CreateMetricRequest is the body for POST /v1/metrics.
type CreateMetricRequest struct {
	// Name is a human-readable label (e.g. "API Calls").
	Name string `json:"name"`
	// EventName is the event_name field value that this metric tracks.
	EventName string `json:"event_name"`
	// Aggregation determines how events are counted.
	// Use the AggregationXxx constants: count, sum, max, minimum, average, unique.
	Aggregation string `json:"aggregation"`
	// Description is optional documentation.
	Description string `json:"description,omitempty"`
	// AggregationProperty is the Properties key whose value is used for
	// sum/max/min/average aggregations.
	AggregationProperty string `json:"aggregation_property,omitempty"`
}

// UpdateMetricRequest is the body for PUT /v1/metrics/{id}.
type UpdateMetricRequest struct {
	Name                string `json:"name,omitempty"`
	EventName           string `json:"event_name,omitempty"`
	Aggregation         string `json:"aggregation,omitempty"`
	Description         string `json:"description,omitempty"`
	AggregationProperty string `json:"aggregation_property,omitempty"`
}

// ListMetricsResponse is returned by GET /v1/metrics.
type ListMetricsResponse struct {
	Metrics []Metric `json:"metrics"`
	Count   int      `json:"count"`
}

// ---------------------------------------------------------------------------
// Plan / Price types
// ---------------------------------------------------------------------------

// PriceTier defines one step in a tiered pricing model.
// Used with PricingModelTiered — pass a []PriceTier marshalled to JSON in
// CreatePriceRequest.Tiers.
type PriceTier struct {
	// UpTo is the upper boundary of this tier (inclusive). A nil value means
	// "infinity" — this tier applies to all remaining usage.
	UpTo *int64 `json:"up_to"`
	// UnitAmount is the price per unit in this tier, expressed as a decimal
	// string (e.g. "0.50", "2.000000").
	UnitAmount string `json:"unit_amount"`
}

// PackageConfig is the price configuration for PricingModelPackage.
// Marshal this struct to JSON and set it as CreatePriceRequest.Tiers.
type PackageConfig struct {
	// PackageSize is the number of units per bundle.
	PackageSize int64 `json:"package_size"`
	// PackagePrice is the price per complete bundle, as a 6-decimal string.
	PackagePrice string `json:"package_price"`
	// RoundUpPartialBlock controls whether partial bundles are rounded up
	// (true) or down/truncated (false). Defaults to true.
	RoundUpPartialBlock bool `json:"round_up_partial_block"`
}

// OverageConfig is the price configuration for PricingModelOverage.
// Marshal this struct to JSON and set it as CreatePriceRequest.Tiers.
type OverageConfig struct {
	// IncludedUnits is the free quota covered by BasePrice.
	// Set to 0 for a pure per-unit overage with no included allowance.
	IncludedUnits int64 `json:"included_units"`
	// BasePrice is the flat fee charged for usage up to IncludedUnits,
	// expressed as a 6-decimal string (e.g. "50.000000").
	// Set to "0.000000" when there is no base fee.
	BasePrice string `json:"base_price"`
	// OveragePrice is the per-unit rate applied to every unit above
	// IncludedUnits, expressed as a 6-decimal string (e.g. "1.500000").
	OveragePrice string `json:"overage_price"`
}

// CreatePriceRequest describes one price to attach to a plan.
type CreatePriceRequest struct {
	// MetricID is the UUID of the metric this price is based on.
	MetricID string `json:"metric_id"`
	// Model is the pricing model. Use PricingModelXxx constants.
	Model string `json:"model"`
	// UnitPrice is the flat price per unit for PricingModelFlat / PricingModelPerUnit.
	// Express as a 6-decimal string, e.g. "2.500000".
	UnitPrice string `json:"unit_price,omitempty"`
	// Tiers holds the model-specific configuration encoded as JSON:
	//   • PricingModelTiered  → json.Marshal([]PriceTier{...})
	//   • PricingModelPackage → json.Marshal(PackageConfig{...})
	//   • PricingModelOverage → json.Marshal(OverageConfig{...})
	Tiers json.RawMessage `json:"tiers,omitempty"`
}

// UpdatePriceRequest describes an updated price for a plan.
type UpdatePriceRequest struct {
	// ID is the UUID of the price to update. Omit to add a new price.
	ID        string          `json:"id,omitempty"`
	MetricID  string          `json:"metric_id,omitempty"`
	Model     string          `json:"model,omitempty"`
	UnitPrice string          `json:"unit_price,omitempty"`
	Tiers     json.RawMessage `json:"tiers,omitempty"`
}

// Price is a pricing rule attached to a plan.
type Price struct {
	ID        string          `json:"id"`
	PlanID    string          `json:"plan_id"`
	MetricID  string          `json:"metric_id"`
	Model     string          `json:"model"`
	UnitPrice string          `json:"unit_price"`
	Tiers     json.RawMessage `json:"tiers,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Plan is a billing plan that defines pricing for one or more metrics.
type Plan struct {
	ID              string    `json:"id"`
	OrgID           string    `json:"org_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	Currency        string    `json:"currency"`
	PlanType        string    `json:"plan_type"`
	BillingPeriod   string    `json:"billing_period"`
	TrialPeriodDays int32     `json:"trial_period_days"`
	Prices          []Price   `json:"prices,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreatePlanRequest is the body for POST /v1/plans.
type CreatePlanRequest struct {
	// Name is the plan's display name.
	Name string `json:"name"`
	// Description is optional documentation.
	Description string `json:"description,omitempty"`
	// Currency is the ISO 4217 currency code. Defaults to "NGN".
	Currency string `json:"currency,omitempty"`
	// PlanType is either "collection" (billing customers) or "payout" (paying out to vendors).
	// Defaults to "collection".
	PlanType string `json:"plan_type,omitempty"`
	// BillingPeriod controls the invoice cadence. Use BillingPeriodXxx constants.
	// Defaults to "monthly".
	BillingPeriod string `json:"billing_period,omitempty"`
	// Prices is an optional list of pricing rules to attach immediately.
	Prices []CreatePriceRequest `json:"prices,omitempty"`
}

// UpdatePlanRequest is the body for PUT /v1/plans/{id}.
type UpdatePlanRequest struct {
	Name          string               `json:"name,omitempty"`
	Description   string               `json:"description,omitempty"`
	Currency      string               `json:"currency,omitempty"`
	PlanType      string               `json:"plan_type,omitempty"`
	BillingPeriod string               `json:"billing_period,omitempty"`
	Prices        []UpdatePriceRequest `json:"prices,omitempty"`
}

// ListPlansResponse is returned by GET /v1/plans.
type ListPlansResponse struct {
	Plans []Plan `json:"plans"`
	Count int    `json:"count"`
}

// ---------------------------------------------------------------------------
// Subscription types
// ---------------------------------------------------------------------------

// Subscription links a customer to a billing plan.
type Subscription struct {
	ID                 string     `json:"id"`
	OrgID              string     `json:"org_id"`
	CustomerID         string     `json:"customer_id"`
	PlanID             string     `json:"plan_id"`
	Status             string     `json:"status"`
	CurrentPeriodStart time.Time  `json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end"`
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateSubscriptionRequest is the body for POST /v1/subscriptions.
type CreateSubscriptionRequest struct {
	// CustomerID is the UUID of the customer to subscribe.
	CustomerID string `json:"customer_id"`
	// PlanID is the UUID of the plan to subscribe the customer to.
	PlanID string `json:"plan_id"`
}

// ListSubscriptionsParams are the optional query parameters for GET /v1/subscriptions.
type ListSubscriptionsParams struct {
	// CustomerID filters subscriptions to a specific customer.
	CustomerID string
	// PlanID filters subscriptions to a specific plan.
	PlanID string
	// Status filters by subscription status (active, paused, canceled).
	Status string
}

// ListSubscriptionsResponse is returned by GET /v1/subscriptions.
type ListSubscriptionsResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
	Count         int            `json:"count"`
}

// ---------------------------------------------------------------------------
// Payout account types
// ---------------------------------------------------------------------------

// PayoutAccount is a bank or mobile-money account that a customer can be paid to.
type PayoutAccount struct {
	ID                string          `json:"id"`
	CustomerID        string          `json:"customer_id"`
	OrgID             string          `json:"org_id"`
	AccountName       string          `json:"account_name"`
	BankName          string          `json:"bank_name,omitempty"`
	BankCode          string          `json:"bank_code,omitempty"`
	AccountNumber     string          `json:"account_number,omitempty"`
	MobileMoneyNumber string          `json:"mobile_money_number,omitempty"`
	PayoutMethod      string          `json:"payout_method"`
	Currency          string          `json:"currency"`
	IsDefault         bool            `json:"is_default"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// CreatePayoutAccountRequest is the body for POST /v1/customers/{id}/payout-accounts.
type CreatePayoutAccountRequest struct {
	// AccountName is the name on the account.
	AccountName string `json:"account_name"`
	// PayoutMethod is either "bank_transfer" or "mobile_money".
	PayoutMethod      string          `json:"payout_method"`
	BankName          string          `json:"bank_name,omitempty"`
	BankCode          string          `json:"bank_code,omitempty"`
	AccountNumber     string          `json:"account_number,omitempty"`
	MobileMoneyNumber string          `json:"mobile_money_number,omitempty"`
	Currency          string          `json:"currency,omitempty"`
	IsDefault         bool            `json:"is_default,omitempty"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

// UpdatePayoutAccountRequest is the body for PUT /v1/customers/{id}/payout-accounts/{account_id}.
type UpdatePayoutAccountRequest struct {
	AccountName       string          `json:"account_name,omitempty"`
	PayoutMethod      string          `json:"payout_method,omitempty"`
	BankName          string          `json:"bank_name,omitempty"`
	AccountNumber     string          `json:"account_number,omitempty"`
	Currency          string          `json:"currency,omitempty"`
	IsDefault         bool            `json:"is_default,omitempty"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

// ListPayoutAccountsResponse is returned by GET /v1/customers/{id}/payout-accounts.
type ListPayoutAccountsResponse struct {
	PayoutAccounts []PayoutAccount `json:"payout_accounts"`
	Count          int             `json:"count"`
}

// ---------------------------------------------------------------------------
// Invoice types
// ---------------------------------------------------------------------------

// InvoiceLineItem is one line on an invoice showing usage of a single metric.
type InvoiceLineItem struct {
	ID          string    `json:"id"`
	InvoiceID   string    `json:"invoice_id"`
	MetricID    string    `json:"metric_id"`
	PriceID     string    `json:"price_id,omitempty"`
	Description string    `json:"description"`
	Quantity    string    `json:"quantity"`
	UnitPrice   string    `json:"unit_price"`
	Amount      string    `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}

// Invoice represents a billing invoice.
// All monetary values are decimal strings (e.g. "1500.00") to avoid
// floating-point precision issues.
type Invoice struct {
	ID                string            `json:"id"`
	OrgID             string            `json:"org_id"`
	CustomerID        string            `json:"customer_id"`
	SubscriptionID    string            `json:"subscription_id"`
	Status            string            `json:"status"`
	Currency          string            `json:"currency"`
	Subtotal          string            `json:"subtotal"`
	VATEnabled        bool              `json:"vat_enabled"`
	VATRate           string            `json:"vat_rate,omitempty"`
	VATAmount         string            `json:"vat_amount,omitempty"`
	Total             string            `json:"total"`
	PeriodStart       time.Time         `json:"period_start"`
	PeriodEnd         time.Time         `json:"period_end"`
	FinalizedAt       *time.Time        `json:"finalized_at,omitempty"`
	PaidAt            *time.Time        `json:"paid_at,omitempty"`
	ProviderInvoiceID string            `json:"provider_invoice_id,omitempty"`
	LineItems         []InvoiceLineItem `json:"line_items,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// GenerateInvoiceRequest is the body for POST /v1/invoices/generate.
type GenerateInvoiceRequest struct {
	// SubscriptionID is the UUID of the subscription to generate an invoice for.
	SubscriptionID string `json:"subscription_id"`
}

// ListInvoicesParams are optional query parameters for GET /v1/invoices.
type ListInvoicesParams struct {
	// Status filters by invoice status (draft, finalized, paid, void).
	Status string
	// CustomerID filters invoices to a specific customer.
	CustomerID string
}

// ListInvoicesResponse is returned by GET /v1/invoices.
type ListInvoicesResponse struct {
	Invoices []Invoice `json:"invoices"`
	Count    int       `json:"count"`
}

// ---------------------------------------------------------------------------
// Usage types
// ---------------------------------------------------------------------------

// UsageParams are the optional query parameters for GET /v1/usage.
type UsageParams struct {
	// CustomerID filters rollups to a specific customer.
	CustomerID string
	// MetricID filters rollups to a specific metric (UUID).
	MetricID string
	// From is the lower bound of the period_start to query (RFC3339).
	// Defaults to the start of the current billing period.
	From *time.Time
	// To is the exclusive upper bound of the period_start to query (RFC3339).
	// Defaults to the end of the current billing period.
	To *time.Time
}

// UsageRollup is one aggregated usage record for a customer/metric/period tuple.
type UsageRollup struct {
	ID          string     `json:"id"`
	OrgID       string     `json:"org_id"`
	CustomerID  string     `json:"customer_id"`
	MetricID    string     `json:"metric_id"`
	PeriodStart time.Time  `json:"period_start"`
	PeriodEnd   time.Time  `json:"period_end"`
	Aggregation string     `json:"aggregation"`
	// Value is the aggregated usage (count, sum, max, etc.).
	Value       float64    `json:"value"`
	EventCount  int64      `json:"event_count"`
	LastEventAt *time.Time `json:"last_event_at,omitempty"`
	IsTest      bool       `json:"is_test"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UsageQueryResult is returned by GET /v1/usage.
type UsageQueryResult struct {
	Rollups []UsageRollup `json:"rollups"`
	Count   int           `json:"count"`
}

// ---------------------------------------------------------------------------
// Portal token types
// ---------------------------------------------------------------------------

// PortalToken is a single-use shareable link that grants a customer read-only
// access to their invoices, payout slips, subscriptions, and payout accounts
// in the Monigo hosted portal.
type PortalToken struct {
	ID         string     `json:"id"`
	OrgID      string     `json:"org_id"`
	CustomerID string     `json:"customer_id"`
	// Token is the opaque 64-character hex string embedded in the portal URL.
	Token      string     `json:"token"`
	Label      string     `json:"label"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	// PortalURL is the fully-qualified URL to share with the customer.
	// Example: https://app.monigo.co/portal/<token>
	PortalURL string `json:"portal_url"`
}

// CreatePortalTokenRequest is the body for POST /v1/portal/tokens.
type CreatePortalTokenRequest struct {
	// CustomerExternalID is the external_id you assigned this customer when
	// you called Customers.Create.
	CustomerExternalID string `json:"customer_external_id"`
	// Label is an optional human-readable name for this link
	// (e.g. "Main portal link").
	Label string `json:"label,omitempty"`
	// ExpiresAt is an optional RFC3339 timestamp after which the token is
	// automatically rejected. Omit for a permanent link.
	ExpiresAt string `json:"expires_at,omitempty"`
}

// ListPortalTokensResponse is returned by GET /v1/portal/tokens.
type ListPortalTokensResponse struct {
	Tokens []PortalToken `json:"tokens"`
	Count  int           `json:"count"`
}

// ---------------------------------------------------------------------------
// Event replay types
// ---------------------------------------------------------------------------

// EventReplayJob tracks the progress of an event replay operation.
type EventReplayJob struct {
	ID             string     `json:"id"`
	OrgID          string     `json:"org_id"`
	InitiatedBy    string     `json:"initiated_by"`
	Status         string     `json:"status"`
	FromTimestamp  time.Time  `json:"from_timestamp"`
	ToTimestamp    time.Time  `json:"to_timestamp"`
	EventName      *string    `json:"event_name,omitempty"`
	IsTest         bool       `json:"is_test"`
	EventsTotal    int64      `json:"events_total"`
	EventsReplayed int64      `json:"events_replayed"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
