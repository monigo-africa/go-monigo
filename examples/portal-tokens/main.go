// Package main demonstrates Monigo customer portal token management:
//
//  1. Create a permanent portal link for a customer
//  2. Create a time-limited portal link (expires in 30 days)
//  3. List all active portal tokens for the customer
//  4. Revoke the time-limited token
//  5. Verify only the permanent link remains
//
// Run:
//
//	MONIGO_API_KEY=sk_test_... CUSTOMER_EXTERNAL_ID=usr_abc123 go run ./examples/portal-tokens
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
	externalID := os.Getenv("CUSTOMER_EXTERNAL_ID")
	if externalID == "" {
		log.Fatal("CUSTOMER_EXTERNAL_ID environment variable is required")
	}

	opts := []monigo.Option{}
	if baseURL := os.Getenv("MONIGO_BASE_URL"); baseURL != "" {
		opts = append(opts, monigo.WithBaseURL(baseURL))
	}

	client := monigo.New(apiKey, opts...)
	ctx := context.Background()

	// -----------------------------------------------------------------------
	// 1. Create a permanent portal link
	// -----------------------------------------------------------------------
	fmt.Printf("→ Creating permanent portal link for customer %q...\n", externalID)
	permanent, err := client.PortalTokens.Create(ctx, monigo.CreatePortalTokenRequest{
		CustomerExternalID: externalID,
		Label:              "Main portal link",
	})
	if err != nil {
		log.Fatalf("create permanent token: %v", err)
	}
	fmt.Printf("  ✓ Token ID:  %s\n", permanent.ID)
	fmt.Printf("  ✓ Portal URL: %s\n", permanent.PortalURL)

	// -----------------------------------------------------------------------
	// 2. Create a time-limited portal link (expires in 30 days)
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Creating 30-day portal link...")
	expiry := time.Now().Add(30 * 24 * time.Hour)
	timed, err := client.PortalTokens.Create(ctx, monigo.CreatePortalTokenRequest{
		CustomerExternalID: externalID,
		Label:              "30-day invoice link",
		ExpiresAt:          expiry.UTC().Format(time.RFC3339),
	})
	if err != nil {
		log.Fatalf("create timed token: %v", err)
	}
	fmt.Printf("  ✓ Token ID:   %s\n", timed.ID)
	fmt.Printf("  ✓ Expires at: %s\n", expiry.Format("2006-01-02"))
	fmt.Printf("  ✓ Portal URL: %s\n", timed.PortalURL)

	// -----------------------------------------------------------------------
	// 3. List all tokens for the customer
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Listing all portal tokens...")
	resp, err := client.PortalTokens.List(ctx, externalID)
	if err != nil {
		log.Fatalf("list tokens: %v", err)
	}
	fmt.Printf("  Found %d token(s):\n", resp.Count)
	for _, tok := range resp.Tokens {
		expStr := "never"
		if tok.ExpiresAt != nil {
			expStr = tok.ExpiresAt.Format("2006-01-02")
		}
		fmt.Printf("    • [%s] %q  expires=%s\n", tok.ID, tok.Label, expStr)
	}

	// -----------------------------------------------------------------------
	// 4. Revoke the time-limited token
	// -----------------------------------------------------------------------
	fmt.Printf("\n→ Revoking timed token %s...\n", timed.ID)
	if err := client.PortalTokens.Revoke(ctx, timed.ID); err != nil {
		log.Fatalf("revoke token: %v", err)
	}
	fmt.Println("  ✓ Token revoked — that portal URL will now return 401")

	// -----------------------------------------------------------------------
	// 5. Re-list to confirm only the permanent token remains
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Re-listing tokens after revocation...")
	resp, err = client.PortalTokens.List(ctx, externalID)
	if err != nil {
		log.Fatalf("list tokens (after revoke): %v", err)
	}
	fmt.Printf("  Active token(s): %d\n", resp.Count)
	for _, tok := range resp.Tokens {
		fmt.Printf("    • [%s] %q\n", tok.ID, tok.Label)
	}

	// -----------------------------------------------------------------------
	// Summary
	// -----------------------------------------------------------------------
	fmt.Println()
	fmt.Println("✅ Portal token management complete!")
	fmt.Println()
	fmt.Println("Share the permanent portal URL with your customer:")
	fmt.Printf("   %s\n", permanent.PortalURL)
	fmt.Println()
	fmt.Println("Customers can use this URL to view their invoices, payout slips,")
	fmt.Println("subscriptions, and payout accounts without needing a Monigo account.")
}
