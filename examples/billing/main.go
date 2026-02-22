// Package main demonstrates the Monigo invoice lifecycle:
//
//  1. Generate a draft invoice for a subscription
//  2. Print the line items
//  3. Finalize the invoice
//  4. (Optionally) void it
//
// Run:
//
//	MONIGO_API_KEY=sk_test_... SUBSCRIPTION_ID=<uuid> go run ./examples/billing
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	monigo "github.com/monigo-africa/go-monigo"
)

func main() {
	apiKey := os.Getenv("MONIGO_API_KEY")
	if apiKey == "" {
		log.Fatal("MONIGO_API_KEY environment variable is required")
	}
	subscriptionID := os.Getenv("SUBSCRIPTION_ID")
	if subscriptionID == "" {
		log.Fatal("SUBSCRIPTION_ID environment variable is required")
	}

	opts := []monigo.Option{}
	if baseURL := os.Getenv("MONIGO_BASE_URL"); baseURL != "" {
		opts = append(opts, monigo.WithBaseURL(baseURL))
	}

	client := monigo.New(apiKey, opts...)
	ctx := context.Background()

	// -----------------------------------------------------------------------
	// 1. Generate a draft invoice
	// -----------------------------------------------------------------------
	fmt.Println("→ Generating draft invoice...")
	invoice, err := client.Invoices.Generate(ctx, subscriptionID)
	if err != nil {
		log.Fatalf("generate invoice: %v", err)
	}
	printInvoice(invoice)

	// -----------------------------------------------------------------------
	// 2. List all invoices for this customer to confirm it appears
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Listing invoices for customer", invoice.CustomerID)
	list, err := client.Invoices.List(ctx, monigo.ListInvoicesParams{
		CustomerID: invoice.CustomerID,
	})
	if err != nil {
		log.Fatalf("list invoices: %v", err)
	}
	fmt.Printf("  Found %d invoice(s)\n", list.Count)

	// -----------------------------------------------------------------------
	// 3. Finalize the invoice
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Finalizing invoice...")
	finalized, err := client.Invoices.Finalize(ctx, invoice.ID)
	if err != nil {
		log.Fatalf("finalize invoice: %v", err)
	}
	fmt.Printf("  ✓ Invoice status: %s (was %s)\n", finalized.Status, invoice.Status)

	// -----------------------------------------------------------------------
	// Optional: void the invoice (comment out to keep it finalized)
	// -----------------------------------------------------------------------
	if os.Getenv("VOID_INVOICE") == "true" {
		fmt.Println("\n→ Voiding invoice...")
		voided, err := client.Invoices.Void(ctx, finalized.ID)
		if err != nil {
			log.Fatalf("void invoice: %v", err)
		}
		fmt.Printf("  ✓ Invoice status: %s\n", voided.Status)
	}

	fmt.Println("\n✅ Billing example complete!")
}

func printInvoice(inv *monigo.Invoice) {
	fmt.Printf("\n  Invoice ID:    %s\n", inv.ID)
	fmt.Printf("  Customer:      %s\n", inv.CustomerID)
	fmt.Printf("  Subscription:  %s\n", inv.SubscriptionID)
	fmt.Printf("  Period:        %s → %s\n",
		inv.PeriodStart.Format("2006-01-02"),
		inv.PeriodEnd.Format("2006-01-02"))
	fmt.Printf("  Status:        %s\n", inv.Status)
	fmt.Printf("  Currency:      %s\n", inv.Currency)
	fmt.Printf("  Subtotal:      %s\n", inv.Subtotal)
	fmt.Printf("  Total:         %s\n", inv.Total)

	if len(inv.LineItems) > 0 {
		fmt.Println("\n  Line Items:")
		for _, li := range inv.LineItems {
			fmt.Printf("    %-40s  qty=%-10s  unit=%-10s  amount=%s\n",
				li.Description, li.Quantity, li.UnitPrice, li.Amount)
		}
	} else {
		fmt.Println("  (No line items yet — usage may not be rolled up yet)")
	}
}
