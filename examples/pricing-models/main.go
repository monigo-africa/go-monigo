// Package main demonstrates every pricing model supported by Monigo.
//
// Six plans are created, each using a different pricing model, all billed
// monthly in NGN.  A single customer is subscribed to each plan so you can
// inspect the resulting structure in the dashboard.
//
// Pricing models covered:
//
//	flat           – fixed price per unit, no tiers (e.g. ₦2 per API call)
//	tiered         – graduated tiers; each unit is charged at the rate of the
//	                 tier it falls in (first N units at price A, next M at B…)
//	volume         – whole usage is charged at the rate of the highest tier
//	                 reached (one rate applies to every unit)
//	package        – charge per block/bundle of N units (e.g. ₦500 per 1 000 SMS)
//	overage        – free up to an included quota, then a per-unit rate beyond it
//	weighted_tiered – like tiered, but the blended average across all tiers is
//	                  used for the final charge
//
// Run:
//
//	MONIGO_API_KEY=sk_test_... go run ./examples/pricing-models
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	monigo "github.com/monigo-africa/go-monigo"
)

func ptr[T any](v T) *T { return &v }

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

	storageGBMetric, err := client.Metrics.Create(ctx, monigo.CreateMetricRequest{
		Name:                "Storage (GB)",
		EventName:           "storage_write",
		Aggregation:         monigo.AggregationSum,
		AggregationProperty: "gb",
		Description:         "Total gigabytes written",
	})
	if err != nil {
		log.Fatalf("create storage metric: %v", err)
	}
	fmt.Printf("  ✓ Metric: %s (%s)\n", storageGBMetric.Name, storageGBMetric.ID)

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
	// 1. Flat pricing
	//
	// Every unit costs exactly the same fixed amount.
	// Use case: simple per-call billing where rate never changes.
	//
	//  0 – ∞  calls  →  ₦2.00 each
	// -----------------------------------------------------------------------
	fmt.Println("→ [1/6] Creating FLAT pricing plan...")
	flatPlan, err := client.Plans.Create(ctx, monigo.CreatePlanRequest{
		Name:          "Flat – API Calls",
		Description:   "₦2.00 per API call, no tiers.",
		Currency:      "NGN",
		PlanType:      monigo.PlanTypeCollection,
		BillingPeriod: monigo.BillingPeriodMonthly,
		Prices: []monigo.CreatePriceRequest{
			{
				MetricID:  apiCallMetric.ID,
				Model:     monigo.PricingModelFlat,
				UnitPrice: "2.000000", // ₦2 per call
			},
		},
	})
	if err != nil {
		log.Fatalf("create flat plan: %v", err)
	}
	printPlan(flatPlan)

	// -----------------------------------------------------------------------
	// 2. Tiered pricing (graduated)
	//
	// Each unit is charged at the rate of the tier it falls into.
	// Heavy usage is progressively cheaper per unit.
	//
	//    1 –  1 000  calls  →  ₦5.00 each
	// 1 001 – 10 000  calls  →  ₦3.00 each
	// 10 001+          calls  →  ₦1.00 each
	// -----------------------------------------------------------------------
	fmt.Println("→ [2/6] Creating TIERED (graduated) pricing plan...")
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
				Tiers: []monigo.PriceTier{
					{UpTo: ptr[int64](1_000), UnitAmount: "5.000000"},  // first 1 000: ₦5 each
					{UpTo: ptr[int64](10_000), UnitAmount: "3.000000"}, // next 9 000: ₦3 each
					{UpTo: nil, UnitAmount: "1.000000"},                 // beyond 10 000: ₦1 each
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("create tiered plan: %v", err)
	}
	printPlan(tieredPlan)

	// -----------------------------------------------------------------------
	// 3. Volume pricing
	//
	// The customer's total usage determines which tier they land in, and that
	// single rate is applied to ALL units — not just the units in that tier.
	// Contrast with tiered where each tier's rate applies only to units within it.
	//
	//  0 –  5 000  GB  →  ₦10.00 / GB  (applied to every GB if ≤ 5 000)
	//  5 001 – 20 000 GB  →  ₦7.00  / GB  (applied to every GB if 5 001–20 000)
	//  20 001+          GB  →  ₦5.00  / GB  (applied to every GB if > 20 000)
	// -----------------------------------------------------------------------
	fmt.Println("→ [3/6] Creating VOLUME pricing plan...")
	volumePlan, err := client.Plans.Create(ctx, monigo.CreatePlanRequest{
		Name:          "Volume – Storage",
		Description:   "One rate for all storage, based on total usage tier.",
		Currency:      "NGN",
		PlanType:      monigo.PlanTypeCollection,
		BillingPeriod: monigo.BillingPeriodMonthly,
		Prices: []monigo.CreatePriceRequest{
			{
				MetricID: storageGBMetric.ID,
				Model:    monigo.PricingModelVolume,
				Tiers: []monigo.PriceTier{
					{UpTo: ptr[int64](5_000), UnitAmount: "10.000000"},  // ≤ 5 000 GB: ₦10/GB all
					{UpTo: ptr[int64](20_000), UnitAmount: "7.000000"},  // ≤ 20 000 GB: ₦7/GB all
					{UpTo: nil, UnitAmount: "5.000000"},                  // > 20 000 GB: ₦5/GB all
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("create volume plan: %v", err)
	}
	printPlan(volumePlan)

	// -----------------------------------------------------------------------
	// 4. Package pricing
	//
	// Usage is sold in fixed-size bundles (packages).  The customer is charged
	// for whole packages, rounding up any partial package.
	//
	//  1 package = 1 000 SMS  →  ₦500 per package
	//
	// A customer sending 1 500 SMS is charged for 2 packages = ₦1 000.
	// -----------------------------------------------------------------------
	fmt.Println("→ [4/6] Creating PACKAGE pricing plan...")
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
				// UnitPrice is the price per package.
				// The package size (1 000 SMS) is configured on the metric or plan
				// in the Monigo dashboard; the SDK passes the per-package price here.
				UnitPrice: "500.000000", // ₦500 per bundle of 1 000 SMS
			},
		},
	})
	if err != nil {
		log.Fatalf("create package plan: %v", err)
	}
	printPlan(packagePlan)

	// -----------------------------------------------------------------------
	// 5. Overage pricing
	//
	// A free included quota is bundled into the plan; usage beyond that
	// threshold is billed at a per-unit overage rate.
	//
	//  0 – 10 000  calls/month  →  included (₦0)
	//  10 001+      calls/month  →  ₦1.50 each
	//
	// The first tier's UnitAmount "0.000000" represents the included quota.
	// The last tier (UpTo = nil) is the overage rate.
	// -----------------------------------------------------------------------
	fmt.Println("→ [5/6] Creating OVERAGE pricing plan...")
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
				Tiers: []monigo.PriceTier{
					{UpTo: ptr[int64](10_000), UnitAmount: "0.000000"}, // first 10 000: free
					{UpTo: nil, UnitAmount: "1.500000"},                 // beyond: ₦1.50 each
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("create overage plan: %v", err)
	}
	printPlan(overagePlan)

	// -----------------------------------------------------------------------
	// 6. Weighted tiered pricing
	//
	// Similar to graduated tiered pricing, but the final amount is a weighted
	// average of all tier rates based on how much usage fell into each tier.
	// This produces a single blended per-unit price rather than separate line
	// items per tier.
	//
	//    1 –  1 000  GB  →  ₦8.00 / GB
	// 1 001 –  5 000  GB  →  ₦6.00 / GB
	// 5 001+           GB  →  ₦4.00 / GB
	// -----------------------------------------------------------------------
	fmt.Println("→ [6/6] Creating WEIGHTED TIERED pricing plan...")
	weightedPlan, err := client.Plans.Create(ctx, monigo.CreatePlanRequest{
		Name:          "Weighted Tiered – Storage",
		Description:   "Blended per-GB rate derived from weighted average across tiers.",
		Currency:      "NGN",
		PlanType:      monigo.PlanTypeCollection,
		BillingPeriod: monigo.BillingPeriodMonthly,
		Prices: []monigo.CreatePriceRequest{
			{
				MetricID: storageGBMetric.ID,
				Model:    monigo.PricingModelWeightedTiered,
				Tiers: []monigo.PriceTier{
					{UpTo: ptr[int64](1_000), UnitAmount: "8.000000"},  // first 1 000 GB
					{UpTo: ptr[int64](5_000), UnitAmount: "6.000000"},  // next 4 000 GB
					{UpTo: nil, UnitAmount: "4.000000"},                 // beyond 5 000 GB
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("create weighted tiered plan: %v", err)
	}
	printPlan(weightedPlan)

	// -----------------------------------------------------------------------
	// Subscribe the demo customer to every plan so it's all queryable
	// -----------------------------------------------------------------------
	fmt.Println("→ Subscribing customer to all plans...")
	plans := []*monigo.Plan{flatPlan, tieredPlan, volumePlan, packagePlan, overagePlan, weightedPlan}
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
		fmt.Printf("     %-35s  id=%s  model=%s\n", p.Name, p.ID, pricingModelSummary(p))
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
			fmt.Printf("         price id=%-38s  model=%s (tiered)\n",
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
