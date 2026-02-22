// Package main demonstrates querying Monigo usage rollups and printing a
// formatted usage report to the terminal.
//
// Filters (all optional via environment variables):
//   - CUSTOMER_ID — show only one customer
//   - METRIC_ID   — show only one metric
//   - FROM        — RFC3339 start time (default: start of current month)
//   - TO          — RFC3339 end time (default: now)
//
// Run:
//
//	MONIGO_API_KEY=sk_test_... go run ./examples/usage-report
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
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

	params := monigo.UsageParams{
		CustomerID: os.Getenv("CUSTOMER_ID"),
		MetricID:   os.Getenv("METRIC_ID"),
	}

	if v := os.Getenv("FROM"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			log.Fatalf("parse FROM: %v", err)
		}
		params.From = &t
	}
	if v := os.Getenv("TO"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			log.Fatalf("parse TO: %v", err)
		}
		params.To = &t
	}

	result, err := client.Usage.Query(ctx, params)
	if err != nil {
		log.Fatalf("query usage: %v", err)
	}

	if result.Count == 0 {
		fmt.Println("No usage data found for the given filters.")
		return
	}

	fmt.Printf("Usage Report — %d rollup(s)\n\n", result.Count)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CUSTOMER\tMETRIC\tPERIOD\tAGGREGATION\tVALUE\tEVENT COUNT\tTEST")
	fmt.Fprintln(w, "--------\t------\t------\t-----------\t-----\t-----------\t----")

	for _, r := range result.Rollups {
		testFlag := " "
		if r.IsTest {
			testFlag = "✓"
		}
		fmt.Fprintf(w, "%s\t%s\t%s→%s\t%s\t%.4f\t%d\t%s\n",
			truncate(r.CustomerID, 12),
			truncate(r.MetricID, 12),
			r.PeriodStart.Format("01/02"),
			r.PeriodEnd.Format("01/02"),
			r.Aggregation,
			r.Value,
			r.EventCount,
			testFlag,
		)
	}
	w.Flush()

	// Total value per aggregation type
	totals := map[string]float64{}
	for _, r := range result.Rollups {
		totals[r.Aggregation] += r.Value
	}
	fmt.Println()
	fmt.Println("Totals:")
	for agg, total := range totals {
		fmt.Printf("  %s: %.4f\n", agg, total)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
