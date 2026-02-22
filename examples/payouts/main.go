// Package main demonstrates Monigo payout account management and event replay:
//
//  1. Create a payout account for an existing customer
//  2. List all payout accounts for that customer
//  3. Start an event replay for the last 24 hours
//  4. Poll the replay job until it completes
//
// Run:
//
//	MONIGO_API_KEY=sk_test_... CUSTOMER_ID=<uuid> go run ./examples/payouts
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
	customerID := os.Getenv("CUSTOMER_ID")
	if customerID == "" {
		log.Fatal("CUSTOMER_ID environment variable is required")
	}

	opts := []monigo.Option{}
	if baseURL := os.Getenv("MONIGO_BASE_URL"); baseURL != "" {
		opts = append(opts, monigo.WithBaseURL(baseURL))
	}

	client := monigo.New(apiKey, opts...)
	ctx := context.Background()

	// -----------------------------------------------------------------------
	// 1. Create a payout account
	// -----------------------------------------------------------------------
	fmt.Println("→ Creating payout account for customer", customerID)
	account, err := client.PayoutAccounts.Create(ctx, customerID, monigo.CreatePayoutAccountRequest{
		AccountName:   "John Driver",
		PayoutMethod:  monigo.PayoutMethodBankTransfer,
		BankName:      "First Bank Nigeria",
		BankCode:      "011",
		AccountNumber: "3001234567",
		Currency:      "NGN",
		IsDefault:     true,
	})
	if err != nil {
		log.Fatalf("create payout account: %v", err)
	}
	fmt.Printf("  ✓ Account created: %s (%s)\n", account.AccountName, account.ID)

	// -----------------------------------------------------------------------
	// 2. List all payout accounts
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Listing payout accounts...")
	accounts, err := client.PayoutAccounts.List(ctx, customerID)
	if err != nil {
		log.Fatalf("list payout accounts: %v", err)
	}
	fmt.Printf("  Found %d account(s):\n", accounts.Count)
	for _, a := range accounts.PayoutAccounts {
		defaultMark := ""
		if a.IsDefault {
			defaultMark = " (default)"
		}
		fmt.Printf("    • %s — %s %s%s\n", a.AccountName, a.BankName, a.AccountNumber, defaultMark)
	}

	// -----------------------------------------------------------------------
	// 3. Start an event replay for the last 24 hours
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Starting event replay for last 24 hours...")
	to := time.Now().UTC()
	from := to.Add(-24 * time.Hour)
	job, err := client.Events.StartReplay(ctx, from, to, nil)
	if err != nil {
		log.Fatalf("start replay: %v", err)
	}
	fmt.Printf("  ✓ Replay job started: %s (status: %s)\n", job.ID, job.Status)

	// -----------------------------------------------------------------------
	// 4. Poll until complete (with a timeout)
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Polling replay job status...")
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(3 * time.Second)
		updated, err := client.Events.GetReplay(ctx, job.ID)
		if err != nil {
			log.Fatalf("get replay: %v", err)
		}
		fmt.Printf("  Status: %-12s  replayed=%d/%d\n",
			updated.Status, updated.EventsReplayed, updated.EventsTotal)
		if updated.Status == "completed" || updated.Status == "failed" {
			job = updated
			break
		}
	}

	fmt.Printf("\n✅ Replay finished with status: %s\n", job.Status)
	if job.ErrorMessage != nil {
		fmt.Printf("   Error: %s\n", *job.ErrorMessage)
	}
}
