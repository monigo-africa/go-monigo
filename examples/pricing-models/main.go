// Package main demonstrates every pricing model supported by Monigo.
//
// Four plans are created, each using a different pricing model, all billed
// monthly in NGN.  A single customer is subscribed to each plan so you can
// inspect the resulting structure in the dashboard.
//
// Pricing models covered:
//
//	flat_unit  – fixed price per unit (PricingModelFlat / PricingModelPerUnit)
//	tiered     – graduated tiers; each unit is charged at the rate of the tier
//	             it falls into. Requires []PriceTier marshalled into Tiers.
//	package    – charge per bundle of N units. Requires PackageConfig in Tiers.
//	overage    – flat BasePrice covers IncludedUnits; OveragePrice per unit
//	             beyond the quota. Requires OverageConfig in Tiers.
//
// Run:
//
//	MONIGO_API_KEY=sk_test_... go run ./examples/pricing-models
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	monigo "github.com/monigo-africa/go-monigo"
)

func ptr[T any](v T) *T { return &v }

// mustMarshal marshals v to JSON or panics. Used only in examples.
func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("mustMarshal: %v", err))
	}
	return b
}

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
	// Shared customer — subscribed to every demo plan below
	// -----------------------------------------------------------------------
	fmt.Println("→ Creating demo customer...")
	customer, err := client.Customers.Create(ctx, monigo.CreateCustomerRequest{
		ExternalID: "pricing-demo-customer",
		Name:       "Pricing Demo Customer",
		Email:      "pricing-demo@example.com",
	})
	if err != nil {
		log.Fatalf("create customer: %v", err)
	}
	fmt.Printf("  ✓ Customer: %s (%s)\n\n", customer.Name, customer.ID)

	// -----------------------------------------------------------------------
	// Shared metrics
	// -----------------------------------------------------------------------
	fmt.Println("→ Creating metrics...")

	apiCallMetric, err := client.Metrics.Create(ctx, monigo.CreateMetricRequest{
		Name:        "API Calls",
		EventName:   "api_call",
		Aggregation: monigo.AggregationCount,
		Description: "Counts every API call",
	})
	if err != nil {
		log.Fatalf("create api_call metric: %v", err)
	}
	fmt.Printf("  ✓ Metric: %s (%s)\n", apiCallMetric.Name, apiCallMetric.ID)

	smsMetric, err := client.Metrics.Create(ctx, monigo.CreateMetricRequest{
		Name:        "SMS Sent",
		EventName:   "sms_sent",
		Aggregation: monigo.AggregationCount,
		Description: "Counts every SMS dispatched",
	})
	if err != nil {
		log.Fatalf("create sms metric: %v", err)
	}
	fmt.Printf("  ✓ Metric: %s (%s)\n\n", smsMetric.Name, smsMetric.ID)

	// -----------------------------------------------------------------------
	// 1. Flat pricing  (model: "flat_unit")
	//
	// Every unit costs exactly the same fixed amount.
	// Use case: simple per-call billing where the rate never changes.
	//
	//  0 – ∞  calls  →  ₦2.00 each
	// -----------------------------------------------------------------------
	fmt.Println("→ [1/4] Creating FLAT pricing plan...")
	flatPlan, err := client.Plans.Create(ctx, monigo.CreatePlanRequest{
		Name:          "Flat – API Calls",
		Description:   "₦2.00 per API call, no tiers.",
		Currency:      "NGN",
		PlanType:      monigo.PlanTypeCollection,
		BillingPeriod: monigo.BillingPeriodMonthly,
		Prices: []monigo.CreatePriceRequest{
			{
				MetricID:  apiCallMetric.ID,
				Model:     monigo.PricingModelFlat, // "flat_unit"
				UnitPrice: "2.000000",              // ₦2 per call
			},
		},
	})
	if err != nil {
		log.Fatalf("create flat plan: %v", err)
	}
	printPlan(flatPlan)

	// -----------------------------------------------------------------------
	// 2. Tiered pricing  (model: "tiered")
	//
	// Each unit is charged at the rate of the tier it falls into.
	// Heavy usage is progressively cheaper per unit.
	// Pass a []PriceTier marshalled to JSON in the Tiers field.
	//
	//    1 –  1 000  calls  →  ₦5.00 each
	// 1 001 – 10 000  calls  →  ₦3.00 each
	// 10 001+          calls  →  ₦1.00 each
	// -----------------------------------------------------------------------
	fmt.Println("→ [2/4] Creating TIERED (graduated) pricing plan...")
	tieredTiers := mustMarshal([]monigo.PriceTier{
		{UpTo: ptr[int64](1_000), UnitAmount: "5.000000"},  // first 1 000: ₦5 each
		{UpTo: ptr[int64](10_000), UnitAmount: "3.000000"}, // next 9 000: ₦3 each
		{UpTo: nil, UnitAmount: "1.000000"},                 // beyond 10 000: ₦1 each
	})
	tieredPlan, err := client.Plans.Create(ctx, monigo.CreatePlanRequest{
		Name:          "Tiered – API Calls",
		Description:   "Graduated tiers: cheaper as volume grows.",
		Currency:      "NGN",
		PlanType:      monigo.PlanTypeCollection,
		BillingPeriod: monigo.BillingPeriodMonthly,
		Prices: []monigo.CreatePriceRequest{
			{
				MetricID: apiCallMetric.ID,
				Model:    monigo.PricingModelTiered,
				Tiers:    tieredTiers,
			},
		},
	})
	if err != nil {
		log.Fatalf("create tiered plan: %v", err)
	}
	printPlan(tieredPlan)

	// -----------------------------------------------------------------------
	// 3. Package pricing  (model: "package")
	//
	// Usage is sold in fixed-size bundles. Partial bundles are rounded up.
	// Pass a PackageConfig marshalled to JSON in the Tiers field.
	//
	//  1 bundle = 1 000 SMS  →  ₦500 per bundle
	//
	// Sending 1 500 SMS → 2 bundles → ₦1 000.
	// -----------------------------------------------------------------------
	fmt.Println("→ [3/4] Creating PACKAGE pricing plan...")
	packageTiers := mustMarshal(monigo.PackageConfig{
		PackageSize:         1000,         // 1 000 SMS per bundle
		PackagePrice:        "500.000000", // ₦500 per bundle
		RoundUpPartialBlock: true,         // partial bundle rounds up
	})
	packagePlan, err := client.Plans.Create(ctx, monigo.CreatePlanRequest{
		Name:          "Package – SMS Bundle",
		Description:   "₦500 per 1 000 SMS bundle. Partial bundles round up.",
		Currency:      "NGN",
		PlanType:      monigo.PlanTypeCollection,
		BillingPeriod: monigo.BillingPeriodMonthly,
		Prices: []monigo.CreatePriceRequest{
			{
				MetricID: smsMetric.ID,
				Model:    monigo.PricingModelPackage,
				Tiers:    packageTiers,
			},
		},
	})
	if err != nil {
		log.Fatalf("create package plan: %v", err)
	}
	printPlan(packagePlan)

	// -----------------------------------------------------------------------
	// 4. Overage pricing  (model: "overage")
	//
	// A flat BasePrice covers up to IncludedUnits. Every unit above the quota
	// is charged at OveragePrice per unit.
	// Pass an OverageConfig marshalled to JSON in the Tiers field.
	//
	//  0 – 10 000  calls/month  →  ₦0 (no base fee, just a free quota)
	//  10 001+      calls/month  →  ₦1.50 each
	// -----------------------------------------------------------------------
	fmt.Println("→ [4/4] Creating OVERAGE pricing plan...")
	overageTiers := mustMarshal(monigo.OverageConfig{
		IncludedUnits: 10_000,    // first 10 000 calls are free
		BasePrice:     "0.000000", // no flat base fee
		OveragePrice:  "1.500000", // ₦1.50 per call beyond the quota
	})
	overagePlan, err := client.Plans.Create(ctx, monigo.CreatePlanRequest{
		Name:          "Overage – API Calls",
		Description:   "10 000 calls/month included, ₦1.50 per call beyond that.",
		Currency:      "NGN",
		PlanType:      monigo.PlanTypeCollection,
		BillingPeriod: monigo.BillingPeriodMonthly,
		Prices: []monigo.CreatePriceRequest{
			{
				MetricID: apiCallMetric.ID,
				Model:    monigo.PricingModelOverage,
				Tiers:    overageTiers,
			},
		},
	})
	if err != nil {
		log.Fatalf("create overage plan: %v", err)
	}
	printPlan(overagePlan)

	// -----------------------------------------------------------------------
	// Subscribe the demo customer to every plan
	// -----------------------------------------------------------------------
	fmt.Println("→ Subscribing customer to all plans...")
	plans := []*monigo.Plan{flatPlan, tieredPlan, packagePlan, overagePlan}
	for _, p := range plans {
		sub, err := client.Subscriptions.Create(ctx, monigo.CreateSubscriptionRequest{
			CustomerID: customer.ID,
			PlanID:     p.ID,
		})
		if err != nil {
			log.Fatalf("subscribe to plan %s: %v", p.Name, err)
		}
		fmt.Printf("  ✓ Subscribed to %-35s  subscription=%s\n", p.Name, sub.ID)
	}

	// -----------------------------------------------------------------------
	// Summary
	// -----------------------------------------------------------------------
	fmt.Println()
	fmt.Println("✅ Pricing models example complete!")
	fmt.Println()
	fmt.Printf("   Customer:   %s\n", customer.ID)
	fmt.Println()
	fmt.Println("   Plans created:")
	for _, p := range plans {
		model := pricingModelSummary(p)
		fmt.Printf("     %-35s  id=%s  model=%s\n", p.Name, p.ID, model)
	}
	fmt.Println()
	fmt.Println("Ingest usage events and then call client.Invoices.Generate(ctx, subscriptionID)")
	fmt.Println("to see each model produce its own line items.")
}

// printPlan prints a compact summary of a newly-created plan.
func printPlan(p *monigo.Plan) {
	fmt.Printf("  ✓ Plan: %-35s  id=%s\n", p.Name, p.ID)
	for _, price := range p.Prices {
		if price.UnitPrice != "" && price.UnitPrice != "0" {
			fmt.Printf("         price id=%-38s  model=%-15s  unit_price=%s\n",
				price.ID, price.Model, price.UnitPrice)
		} else {
			fmt.Printf("         price id=%-38s  model=%s (config in tiers)\n",
				price.ID, price.Model)
		}
	}
	fmt.Println()
}

// pricingModelSummary returns the model name of the first price on a plan.
func pricingModelSummary(p *monigo.Plan) string {
	if len(p.Prices) == 0 {
		return "—"
	}
	return p.Prices[0].Model
}
