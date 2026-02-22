// Package main demonstrates a complete Monigo integration from start to finish:
//
//  1. Create a customer
//  2. Create a metric (api_calls)
//  3. Create a plan with flat-rate pricing
//  4. Subscribe the customer to the plan
//  5. Ingest a batch of usage events
//
// Run:
//
//	MONIGO_API_KEY=sk_test_... go run ./examples/quickstart
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
	apiKey := os.Getenv("MONIGO_API_KEY")
	if apiKey == "" {
		log.Fatal("MONIGO_API_KEY environment variable is required")
	}

	opts := []monigo.Option{}
	if baseURL := os.Getenv("MONIGO_BASE_URL"); baseURL != "" {
		opts = append(opts, monigo.WithBaseURL(baseURL))
	}

	client := monigo.New(apiKey, opts...)
	ctx := context.Background()

	// -----------------------------------------------------------------------
	// 1. Create a customer
	// -----------------------------------------------------------------------
	fmt.Println("→ Creating customer...")
	customer, err := client.Customers.Create(ctx, monigo.CreateCustomerRequest{
		ExternalID: "acme-corp-001",
		Name:       "Acme Corporation",
		Email:      "billing@acme.example",
	})
	if err != nil {
		log.Fatalf("create customer: %v", err)
	}
	fmt.Printf("  ✓ Customer created: %s (%s)\n", customer.Name, customer.ID)

	// -----------------------------------------------------------------------
	// 2. Create a metric
	// -----------------------------------------------------------------------
	fmt.Println("→ Creating metric...")
	metric, err := client.Metrics.Create(ctx, monigo.CreateMetricRequest{
		Name:        "API Calls",
		EventName:   "api_call",
		Aggregation: monigo.AggregationCount,
		Description: "Counts every API call made by a customer",
	})
	if err != nil {
		log.Fatalf("create metric: %v", err)
	}
	fmt.Printf("  ✓ Metric created: %s (%s)\n", metric.Name, metric.ID)

	// -----------------------------------------------------------------------
	// 3. Create a plan with flat-rate pricing (₦2 per API call)
	// -----------------------------------------------------------------------
	fmt.Println("→ Creating plan...")
	plan, err := client.Plans.Create(ctx, monigo.CreatePlanRequest{
		Name:          "API Pro",
		Description:   "₦2 per API call, billed monthly",
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
	if err != nil {
		log.Fatalf("create plan: %v", err)
	}
	fmt.Printf("  ✓ Plan created: %s (%s)\n", plan.Name, plan.ID)

	// -----------------------------------------------------------------------
	// 4. Subscribe the customer to the plan
	// -----------------------------------------------------------------------
	fmt.Println("→ Creating subscription...")
	sub, err := client.Subscriptions.Create(ctx, monigo.CreateSubscriptionRequest{
		CustomerID: customer.ID,
		PlanID:     plan.ID,
	})
	if err != nil {
		log.Fatalf("create subscription: %v", err)
	}
	fmt.Printf("  ✓ Subscription created: %s (status: %s)\n", sub.ID, sub.Status)

	// -----------------------------------------------------------------------
	// 5. Ingest a batch of usage events
	// -----------------------------------------------------------------------
	fmt.Println("→ Ingesting usage events...")
	now := time.Now().UTC()
	events := make([]monigo.IngestEvent, 10)
	for i := range events {
		events[i] = monigo.IngestEvent{
			EventName:      "api_call",
			CustomerID:     customer.ID,
			IdempotencyKey: fmt.Sprintf("quickstart-run-%d-%d", now.UnixNano(), i),
			Timestamp:      now.Add(time.Duration(-i) * time.Second),
			Properties: map[string]any{
				"endpoint": "/v1/predict",
				"method":   "POST",
			},
		}
	}

	resp, err := client.Events.Ingest(ctx, monigo.IngestRequest{Events: events})
	if err != nil {
		log.Fatalf("ingest events: %v", err)
	}
	fmt.Printf("  ✓ Ingested: %d events, Duplicates: %d\n", len(resp.Ingested), len(resp.Duplicates))

	// -----------------------------------------------------------------------
	// Summary
	// -----------------------------------------------------------------------
	fmt.Println()
	fmt.Println("✅ Quickstart complete!")
	fmt.Printf("   Customer:     %s\n", customer.ID)
	fmt.Printf("   Metric:       %s\n", metric.ID)
	fmt.Printf("   Plan:         %s\n", plan.ID)
	fmt.Printf("   Subscription: %s\n", sub.ID)
	fmt.Println()
	fmt.Println("At the end of the billing period, an invoice will be generated automatically.")
	fmt.Println("You can also generate one manually:")
	fmt.Printf("   client.Invoices.Generate(ctx, %q)\n", sub.ID)
}
