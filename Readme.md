# Monigo Go SDK

Official Go client library for the [Monigo](https://monigo.co) usage-based billing API.

## Installation

```bash
go get github.com/monigo-africa/go-monigo
```

Requires Go 1.21+.

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    monigo "github.com/monigo-africa/go-monigo"
)

func main() {
    client := monigo.New(os.Getenv("MONIGO_API_KEY"))
    ctx := context.Background()

    // Create a customer
    customer, err := client.Customers.Create(ctx, monigo.CreateCustomerRequest{
        ExternalID: "user-001",
        Name:       "Acme Corp",
        Email:      "billing@acme.example",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Ingest a usage event
    _, err = client.Events.Ingest(ctx, monigo.IngestRequest{
        Events: []monigo.IngestEvent{
            {
                EventName:      "api_call",
                CustomerID:     customer.ID,
                IdempotencyKey: "unique-key-001",  // re-sending is safe
                Timestamp:      time.Now(),
                Properties:     map[string]any{"endpoint": "/v1/predict"},
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Customer created and event ingested:", customer.ID)
}
```

---

## Client Configuration

```go
import monigo "github.com/monigo-africa/go-monigo"

// Default — points to https://api.monigo.co
client := monigo.New("sk_live_...")

// Custom base URL (self-hosted or local dev)
client := monigo.New("sk_test_...",
    monigo.WithBaseURL("http://localhost:8000"),
)

// Custom HTTP client (timeouts, proxies, transport)
import "net/http"
import "time"

httpClient := &http.Client{Timeout: 30 * time.Second}
client := monigo.New("sk_test_...",
    monigo.WithHTTPClient(httpClient),
)
```

The API key is sent as `Authorization: Bearer {key}` on every request.

---

## Error Handling

All methods return a typed `*APIError` on HTTP 4xx/5xx responses. Use the
sentinel helpers to check specific error conditions:

```go
customer, err := client.Customers.Get(ctx, "cust-123")
if err != nil {
    switch {
    case monigo.IsNotFound(err):
        fmt.Println("customer does not exist")
    case monigo.IsUnauthorized(err):
        fmt.Println("invalid API key")
    case monigo.IsRateLimited(err):
        fmt.Println("slow down — retry after a short pause")
    case monigo.IsQuotaExceeded(err):
        fmt.Println("event quota exhausted — upgrade plan")
    case monigo.IsConflict(err):
        fmt.Println("subscription already exists")
    case monigo.IsValidationError(err):
        var apiErr *monigo.APIError
        errors.As(err, &apiErr)
        fmt.Printf("validation errors: %v\n", apiErr.Details)
    default:
        log.Fatal(err)
    }
}
```

### APIError fields

```go
type APIError struct {
    StatusCode int               // HTTP status (e.g. 404)
    Message    string            // human-readable description
    Details    map[string]string // field-level validation errors (when present)
}
```

---

## Resources

### Events

#### Ingest usage events

```go
resp, err := client.Events.Ingest(ctx, monigo.IngestRequest{
    Events: []monigo.IngestEvent{
        {
            EventName:      "api_call",
            CustomerID:     "cust-uuid",
            IdempotencyKey: "your-unique-key",  // safe to re-send
            Timestamp:      time.Now(),
            Properties: map[string]any{
                "endpoint": "/v1/classify",
                "model":    "gpt-4",
            },
        },
    },
})
// resp.Ingested   — keys of newly ingested events
// resp.Duplicates — keys already seen (safe, no double-counting)
```

**Idempotency:** use a deterministic key derived from your system's event ID so
that retries and replays are automatically de-duplicated.

**Scopes:** requires an API key with the `ingest` scope.

#### Replay events

```go
// Re-process all events in the last 24 hours
from := time.Now().Add(-24 * time.Hour)
to   := time.Now()

job, err := client.Events.StartReplay(ctx, from, to, nil)
// or filter by event name:
name := "api_call"
job, err = client.Events.StartReplay(ctx, from, to, &name)

// Poll until complete
for {
    time.Sleep(3 * time.Second)
    job, err = client.Events.GetReplay(ctx, job.ID)
    if err != nil { log.Fatal(err) }
    fmt.Printf("status=%s  replayed=%d/%d\n",
        job.Status, job.EventsReplayed, job.EventsTotal)
    if job.Status == "completed" || job.Status == "failed" {
        break
    }
}
```

---

### Customers

```go
// Create
customer, err := client.Customers.Create(ctx, monigo.CreateCustomerRequest{
    ExternalID: "user-001",   // ID in your system
    Name:       "Acme Corp",
    Email:      "billing@acme.example",
})

// List
list, err := client.Customers.List(ctx)
fmt.Printf("%d customers\n", list.Count)

// Get
customer, err := client.Customers.Get(ctx, "cust-uuid")

// Update (only non-empty fields are sent)
customer, err = client.Customers.Update(ctx, "cust-uuid", monigo.UpdateCustomerRequest{
    Name: "Acme Corporation",
})

// Delete
err = client.Customers.Delete(ctx, "cust-uuid")
```

---

### Metrics

Metrics define *what* gets counted and *how* aggregation works.

```go
// Create
metric, err := client.Metrics.Create(ctx, monigo.CreateMetricRequest{
    Name:        "API Calls",
    EventName:   "api_call",          // matches IngestEvent.EventName
    Aggregation: monigo.AggregationCount,
    Description: "Counts every API call",
})

// Sum a numeric property (e.g. bytes transferred)
metric, err = client.Metrics.Create(ctx, monigo.CreateMetricRequest{
    Name:                "Data Transfer",
    EventName:           "data_transfer",
    Aggregation:         monigo.AggregationSum,
    AggregationProperty: "bytes",   // sums event.Properties["bytes"]
})

// List / Get / Update / Delete
list, err  := client.Metrics.List(ctx)
metric, err = client.Metrics.Get(ctx, "metric-uuid")
metric, err = client.Metrics.Update(ctx, "metric-uuid", monigo.UpdateMetricRequest{
    Description: "Updated description",
})
err = client.Metrics.Delete(ctx, "metric-uuid")
```

#### Aggregation types

| Constant | Value | Description |
|---|---|---|
| `monigo.AggregationCount` | `"count"` | Count events |
| `monigo.AggregationSum` | `"sum"` | Sum a numeric property |
| `monigo.AggregationMax` | `"max"` | Maximum value of a property |
| `monigo.AggregationMin` | `"minimum"` | Minimum value of a property |
| `monigo.AggregationAverage` | `"average"` | Average value of a property |
| `monigo.AggregationUnique` | `"unique"` | Count distinct values of a property |

---

### Plans

Plans define pricing rules and the billing cadence.

```go
// Flat-rate plan (₦2 per API call, billed monthly)
plan, err := client.Plans.Create(ctx, monigo.CreatePlanRequest{
    Name:          "API Pro",
    Currency:      "NGN",
    PlanType:      monigo.PlanTypeCollection,
    BillingPeriod: monigo.BillingPeriodMonthly,
    Prices: []monigo.CreatePriceRequest{
        {
            MetricID:  metric.ID,
            Model:     monigo.PricingModelFlat,
            UnitPrice: "2.000000",
        },
    },
})

// Tiered pricing
limit := int64(1000)
plan, err = client.Plans.Create(ctx, monigo.CreatePlanRequest{
    Name:     "Tiered Plan",
    Currency: "NGN",
    Prices: []monigo.CreatePriceRequest{
        {
            MetricID: metric.ID,
            Model:    monigo.PricingModelTiered,
            Tiers: []monigo.PriceTier{
                {UpTo: &limit, UnitAmount: "1.000000"}, // first 1 000 units at ₦1
                {UpTo: nil,    UnitAmount: "0.500000"}, // remaining at ₦0.50
            },
        },
    },
})

// Payout plan (paying drivers per km)
plan, err = client.Plans.Create(ctx, monigo.CreatePlanRequest{
    Name:     "Driver Payouts",
    Currency: "NGN",
    PlanType: monigo.PlanTypePayout,
    Prices: []monigo.CreatePriceRequest{
        {MetricID: kmMetric.ID, Model: monigo.PricingModelFlat, UnitPrice: "500.000000"},
    },
})

// List / Get / Update / Delete
list, err := client.Plans.List(ctx)
plan, err  = client.Plans.Get(ctx, "plan-uuid")
plan, err  = client.Plans.Update(ctx, "plan-uuid", monigo.UpdatePlanRequest{Name: "API Pro v2"})
err        = client.Plans.Delete(ctx, "plan-uuid")
```

#### Plan types

| Constant | Value | Description |
|---|---|---|
| `monigo.PlanTypeCollection` | `"collection"` | Bill your customers (money in) |
| `monigo.PlanTypePayout` | `"payout"` | Pay out to vendors/drivers (money out) |

#### Billing periods

| Constant | Value |
|---|---|
| `monigo.BillingPeriodDaily` | `"daily"` |
| `monigo.BillingPeriodWeekly` | `"weekly"` |
| `monigo.BillingPeriodMonthly` | `"monthly"` |
| `monigo.BillingPeriodQuarterly` | `"quarterly"` |
| `monigo.BillingPeriodAnnually` | `"annually"` |

#### Pricing models

| Constant | Value | Description |
|---|---|---|
| `monigo.PricingModelFlat` | `"flat"` | Fixed price per unit |
| `monigo.PricingModelTiered` | `"tiered"` | Graduated tiers (each tier applies to usage within that band) |
| `monigo.PricingModelVolume` | `"volume"` | Volume pricing (entire quantity billed at the tier rate) |
| `monigo.PricingModelPackage` | `"package"` | Price per block of N units |
| `monigo.PricingModelOverage` | `"overage"` | Included base + per-unit above threshold |
| `monigo.PricingModelWeightedTiered` | `"weighted_tiered"` | Weighted average across tiers |

---

### Subscriptions

```go
// Subscribe a customer to a plan
sub, err := client.Subscriptions.Create(ctx, monigo.CreateSubscriptionRequest{
    CustomerID: customer.ID,
    PlanID:     plan.ID,
})
// Returns 409 Conflict if already subscribed. Use monigo.IsConflict(err) to check.

// List with optional filters
list, err := client.Subscriptions.List(ctx, monigo.ListSubscriptionsParams{
    CustomerID: customer.ID,
    Status:     monigo.SubscriptionStatusActive,
})

// Get
sub, err = client.Subscriptions.Get(ctx, sub.ID)

// Change status
sub, err = client.Subscriptions.UpdateStatus(ctx, sub.ID, monigo.SubscriptionStatusPaused)
sub, err = client.Subscriptions.UpdateStatus(ctx, sub.ID, monigo.SubscriptionStatusActive)
sub, err = client.Subscriptions.UpdateStatus(ctx, sub.ID, monigo.SubscriptionStatusCanceled)

// Delete (cancel and remove)
err = client.Subscriptions.Delete(ctx, sub.ID)
```

#### Subscription statuses

| Constant | Value |
|---|---|
| `monigo.SubscriptionStatusActive` | `"active"` |
| `monigo.SubscriptionStatusPaused` | `"paused"` |
| `monigo.SubscriptionStatusCanceled` | `"canceled"` |

---

### Payout Accounts

Bank or mobile-money accounts associated with a customer, used with `payout` plans.

```go
// Bank transfer account
account, err := client.PayoutAccounts.Create(ctx, customer.ID,
    monigo.CreatePayoutAccountRequest{
        AccountName:   "John Driver",
        PayoutMethod:  monigo.PayoutMethodBankTransfer,
        BankName:      "First Bank Nigeria",
        BankCode:      "011",
        AccountNumber: "3001234567",
        Currency:      "NGN",
        IsDefault:     true,
    },
)

// Mobile money account
account, err = client.PayoutAccounts.Create(ctx, customer.ID,
    monigo.CreatePayoutAccountRequest{
        AccountName:       "Jane Driver",
        PayoutMethod:      monigo.PayoutMethodMobileMoney,
        MobileMoneyNumber: "+2348012345678",
        Currency:          "NGN",
    },
)

// List all accounts for a customer
list, err := client.PayoutAccounts.List(ctx, customer.ID)

// Get / Update / Delete
account, err = client.PayoutAccounts.Get(ctx, customer.ID, account.ID)
account, err = client.PayoutAccounts.Update(ctx, customer.ID, account.ID,
    monigo.UpdatePayoutAccountRequest{AccountName: "Updated Name"},
)
err = client.PayoutAccounts.Delete(ctx, customer.ID, account.ID)
```

#### Payout methods

| Constant | Value |
|---|---|
| `monigo.PayoutMethodBankTransfer` | `"bank_transfer"` |
| `monigo.PayoutMethodMobileMoney` | `"mobile_money"` |

---

### Invoices

```go
// Generate a draft invoice for a subscription
invoice, err := client.Invoices.Generate(ctx, sub.ID)
fmt.Printf("Invoice %s: total=%s %s\n", invoice.ID, invoice.Total, invoice.Currency)

// List invoices
list, err := client.Invoices.List(ctx, monigo.ListInvoicesParams{
    Status:     monigo.InvoiceStatusDraft,
    CustomerID: customer.ID,
})

// Get a single invoice (includes line items)
invoice, err = client.Invoices.Get(ctx, invoice.ID)
for _, li := range invoice.LineItems {
    fmt.Printf("  %s: qty=%s unit=%s amount=%s\n",
        li.Description, li.Quantity, li.UnitPrice, li.Amount)
}

// Finalize — makes the invoice payable; no further edits allowed
invoice, err = client.Invoices.Finalize(ctx, invoice.ID)

// Void — mark as void; no longer payable
invoice, err = client.Invoices.Void(ctx, invoice.ID)
```

> **Note:** All monetary values (`Subtotal`, `Total`, `UnitPrice`, `Amount`) are
> returned as decimal strings (e.g. `"1500.00"`) to preserve precision. Parse
> them with `strconv.ParseFloat` or a decimal library as needed.

#### Invoice statuses

| Constant | Value |
|---|---|
| `monigo.InvoiceStatusDraft` | `"draft"` |
| `monigo.InvoiceStatusFinalized` | `"finalized"` |
| `monigo.InvoiceStatusPaid` | `"paid"` |
| `monigo.InvoiceStatusVoid` | `"void"` |

---

### Usage

Query aggregated usage rollups per customer and metric.

```go
// Full current billing period for all customers and metrics
result, err := client.Usage.Query(ctx, monigo.UsageParams{})

// Filter by customer
result, err = client.Usage.Query(ctx, monigo.UsageParams{
    CustomerID: customer.ID,
})

// Filter by metric and custom date range
from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
to   := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
result, err = client.Usage.Query(ctx, monigo.UsageParams{
    MetricID: metric.ID,
    From:     &from,
    To:       &to,
})

fmt.Printf("%d rollups\n", result.Count)
for _, r := range result.Rollups {
    fmt.Printf("  customer=%s metric=%s period=%s value=%.2f events=%d test=%v\n",
        r.CustomerID, r.MetricID,
        r.PeriodStart.Format("2006-01-02"),
        r.Value, r.EventCount, r.IsTest)
}
```

---

## Test Mode

Use a test-mode API key (`sk_test_...`) to send events without affecting live
billing. Test events are flagged with `IsTest: true` in usage rollups and are
isolated from live data.

```go
// Query test-mode usage only
for _, r := range result.Rollups {
    if r.IsTest {
        fmt.Println("test rollup:", r.Value)
    }
}
```

---

## Example Programs

The `examples/` directory contains self-contained runnable programs:

| Program | Description |
|---|---|
| `examples/quickstart/` | End-to-end: customer → metric → plan → subscription → ingest |
| `examples/metering/` | High-volume idempotent batch ingestion with rate-limit retry |
| `examples/billing/` | Invoice lifecycle: generate → inspect line items → finalize → void |
| `examples/payouts/` | Payout account CRUD + event replay with completion polling |
| `examples/usage-report/` | Query usage rollups and render a terminal table |

Run with:

```bash
# Quickstart
MONIGO_API_KEY=sk_test_... go run ./examples/quickstart

# Metering (high volume)
MONIGO_API_KEY=sk_test_... CUSTOMER_ID=<uuid> BATCH_SIZE=50 TOTAL_EVENTS=1000 \
    go run ./examples/metering

# Billing
MONIGO_API_KEY=sk_test_... SUBSCRIPTION_ID=<uuid> go run ./examples/billing

# Payouts (set VOID_INVOICE=true to also void after finalize)
MONIGO_API_KEY=sk_test_... SUBSCRIPTION_ID=<uuid> VOID_INVOICE=true \
    go run ./examples/billing

# Payout accounts + replay
MONIGO_API_KEY=sk_test_... CUSTOMER_ID=<uuid> go run ./examples/payouts

# Usage report
MONIGO_API_KEY=sk_test_... go run ./examples/usage-report
MONIGO_API_KEY=sk_test_... CUSTOMER_ID=<uuid> METRIC_ID=<uuid> \
    FROM=2026-01-01T00:00:00Z TO=2026-02-01T00:00:00Z \
    go run ./examples/usage-report
```

Override the base URL for a local server:

```bash
MONIGO_API_KEY=sk_test_... MONIGO_BASE_URL=http://localhost:8000 \
    go run ./examples/quickstart
```

---

## Running the Tests

The test suite uses `net/http/httptest` only (stdlib, no external test
dependencies). All API calls are intercepted by an in-process mock HTTP server.

```bash
go test ./...

# Verbose
go test -v ./...

# With race detector
go test -race ./...
```

---

## No External Dependencies

The SDK uses only the Go standard library:

| Package | Purpose |
|---|---|
| `net/http` | HTTP client |
| `encoding/json` | Request/response serialisation |
| `context` | Context propagation |
| `net/url` | Query string encoding |
| `fmt`, `io`, `strings`, `time` | Utilities |

---

## License

MIT © Monigo
