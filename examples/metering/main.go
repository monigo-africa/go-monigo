// Package main demonstrates high-volume idempotent event ingestion with Monigo.
//
// This example:
//   - Sends events in configurable batch sizes
//   - Uses a deterministic idempotency key so re-running is always safe
//   - Prints a tally of ingested vs duplicate events
//   - Works in test mode (events are isolated from live billing)
//
// Run:
//
//	MONIGO_API_KEY=sk_test_... CUSTOMER_ID=<uuid> go run ./examples/metering
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	monigo "github.com/monigo-africa/go-monigo"
)

const (
	defaultBatchSize   = 100
	defaultTotalEvents = 500
)

func main() {
	apiKey := os.Getenv("MONIGO_API_KEY")
	if apiKey == "" {
		log.Fatal("MONIGO_API_KEY environment variable is required")
	}
	customerID := os.Getenv("CUSTOMER_ID")
	if customerID == "" {
		log.Fatal("CUSTOMER_ID environment variable is required")
	}

	batchSize := defaultBatchSize
	if v := os.Getenv("BATCH_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			batchSize = n
		}
	}
	totalEvents := defaultTotalEvents
	if v := os.Getenv("TOTAL_EVENTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			totalEvents = n
		}
	}

	opts := []monigo.Option{}
	if baseURL := os.Getenv("MONIGO_BASE_URL"); baseURL != "" {
		opts = append(opts, monigo.WithBaseURL(baseURL))
	}

	client := monigo.New(apiKey, opts...)
	ctx := context.Background()

	fmt.Printf("Metering example: %d events in batches of %d\n\n", totalEvents, batchSize)

	// Use a stable run ID so re-running this program always produces duplicates
	// instead of double-counting — this is the safe, idempotent pattern.
	runID := os.Getenv("RUN_ID")
	if runID == "" {
		runID = fmt.Sprintf("metering-example-%d", time.Now().Unix())
		fmt.Printf("RUN_ID not set, using %s\n", runID)
		fmt.Println("Re-run with RUN_ID=" + runID + " to see all events reported as duplicates.")
		fmt.Println()
	}

	var totalIngested, totalDuplicates int
	batch := make([]monigo.IngestEvent, 0, batchSize)
	now := time.Now().UTC()

	for i := 0; i < totalEvents; i++ {
		batch = append(batch, monigo.IngestEvent{
			EventName:      "api_call",
			CustomerID:     customerID,
			IdempotencyKey: fmt.Sprintf("%s-event-%d", runID, i),
			Timestamp:      now.Add(-time.Duration(i) * time.Millisecond),
			Properties: map[string]any{
				"endpoint":   fmt.Sprintf("/v1/resource/%d", i%10),
				"latency_ms": 42 + i%30,
			},
		})

		if len(batch) == batchSize || i == totalEvents-1 {
			resp, err := client.Events.Ingest(ctx, monigo.IngestRequest{Events: batch})
			if err != nil {
				if monigo.IsRateLimited(err) {
					log.Println("Rate limited — sleeping 1s before retry")
					time.Sleep(time.Second)
					// retry same batch
					resp, err = client.Events.Ingest(ctx, monigo.IngestRequest{Events: batch})
				}
				if err != nil {
					log.Fatalf("ingest batch starting at event %d: %v", i-len(batch)+1, err)
				}
			}
			totalIngested += len(resp.Ingested)
			totalDuplicates += len(resp.Duplicates)
			fmt.Printf("Batch %4d–%4d: ingested=%d, duplicates=%d\n",
				i-len(batch)+1, i, len(resp.Ingested), len(resp.Duplicates))
			batch = batch[:0]
		}
	}

	fmt.Println()
	fmt.Printf("✅ Done. Total ingested: %d, Total duplicates: %d\n", totalIngested, totalDuplicates)
}
