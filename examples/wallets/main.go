// Package main demonstrates the Monigo wallet lifecycle:
//
//  1. Get or create a wallet for a customer
//  2. Credit the wallet (top-up)
//  3. Check the balance
//  4. Debit the wallet (usage charge)
//  5. List transaction history
//  6. Create a virtual account for automatic top-ups
//  7. List virtual accounts
//
// Run:
//
//	MONIGO_API_KEY=mk_test_... CUSTOMER_ID=<uuid> go run ./examples/wallets
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
	// 1. Get or create a wallet
	// -----------------------------------------------------------------------
	fmt.Println("→ Getting or creating wallet...")
	wallet, err := client.Wallets.GetOrCreate(ctx, monigo.GetOrCreateWalletRequest{
		CustomerID: customerID,
		Currency:   "NGN",
	})
	if err != nil {
		log.Fatalf("get or create wallet: %v", err)
	}
	fmt.Printf("  Wallet ID:  %s\n", wallet.ID)
	fmt.Printf("  Currency:   %s\n", wallet.Currency)
	fmt.Printf("  Balance:    %s\n", wallet.Balance)

	// -----------------------------------------------------------------------
	// 2. Credit the wallet (top-up)
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Crediting wallet with 10,000.00...")
	creditResp, err := client.Wallets.Credit(ctx, wallet.ID, monigo.CreditWalletRequest{
		Amount:         "10000.000000",
		Currency:       "NGN",
		Description:    "Manual top-up via SDK example",
		EntryType:      monigo.WalletEntryTypeDeposit,
		ReferenceType:  "sdk_example",
		ReferenceID:    "example_topup_001",
		IdempotencyKey: "sdk_example_topup_001",
	})
	if err != nil {
		log.Fatalf("credit wallet: %v", err)
	}
	fmt.Printf("  New balance: %s\n", creditResp.Wallet.Balance)
	fmt.Printf("  Ledger entries created: %d\n", len(creditResp.LedgerEntries))
	for _, entry := range creditResp.LedgerEntries {
		fmt.Printf("    %s %s %s — %s\n", entry.Direction, entry.Amount, entry.Currency, entry.Description)
	}

	// -----------------------------------------------------------------------
	// 3. Get wallet to verify balance
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Fetching wallet details...")
	detail, err := client.Wallets.Get(ctx, wallet.ID)
	if err != nil {
		log.Fatalf("get wallet: %v", err)
	}
	fmt.Printf("  Balance:          %s %s\n", detail.Wallet.Balance, detail.Wallet.Currency)
	fmt.Printf("  Reserved balance: %s\n", detail.Wallet.ReservedBalance)
	fmt.Printf("  Virtual accounts: %d\n", len(detail.VirtualAccounts))

	// -----------------------------------------------------------------------
	// 4. Debit the wallet (usage charge)
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Debiting wallet with 2,500.00 (usage charge)...")
	debitResp, err := client.Wallets.Debit(ctx, wallet.ID, monigo.DebitWalletRequest{
		Amount:         "2500.000000",
		Currency:       "NGN",
		Description:    "Usage charge — API calls March 2026",
		EntryType:      monigo.WalletEntryTypeUsage,
		ReferenceType:  "sdk_example",
		ReferenceID:    "example_usage_001",
		IdempotencyKey: "sdk_example_usage_001",
	})
	if err != nil {
		if monigo.IsQuotaExceeded(err) {
			fmt.Println("  ⚠ Insufficient wallet balance!")
		} else {
			log.Fatalf("debit wallet: %v", err)
		}
	} else {
		fmt.Printf("  New balance: %s\n", debitResp.Wallet.Balance)
	}

	// -----------------------------------------------------------------------
	// 5. List transaction history
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Listing transactions...")
	txns, err := client.Wallets.ListTransactions(ctx, wallet.ID, monigo.ListTransactionsParams{
		Limit: 10,
	})
	if err != nil {
		log.Fatalf("list transactions: %v", err)
	}
	fmt.Printf("  Total transactions: %d (showing %d)\n", txns.Total, len(txns.Transactions))
	for _, tx := range txns.Transactions {
		fmt.Printf("    %s %s %s — %s (%s → %s)\n",
			tx.Direction, tx.Amount, tx.Currency,
			tx.Description, tx.BalanceBefore, tx.BalanceAfter)
	}

	// -----------------------------------------------------------------------
	// 6. Create a virtual account (skip if CREATE_VA != "true")
	// -----------------------------------------------------------------------
	if os.Getenv("CREATE_VA") == "true" {
		fmt.Println("\n→ Creating virtual account (Paystack)...")
		va, err := client.Wallets.CreateVirtualAccount(ctx, wallet.ID, monigo.CreateVirtualAccountRequest{
			Provider: monigo.VirtualAccountProviderPaystack,
			Currency: "NGN",
		})
		if err != nil {
			log.Fatalf("create virtual account: %v", err)
		}
		fmt.Printf("  Account Number: %s\n", va.AccountNumber)
		fmt.Printf("  Account Name:   %s\n", va.AccountName)
		fmt.Printf("  Bank:           %s (%s)\n", va.BankName, va.BankCode)
		fmt.Printf("  Provider:       %s\n", va.Provider)
	}

	// -----------------------------------------------------------------------
	// 7. List virtual accounts
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Listing virtual accounts...")
	vaList, err := client.Wallets.ListVirtualAccounts(ctx, wallet.ID)
	if err != nil {
		log.Fatalf("list virtual accounts: %v", err)
	}
	if vaList.Count == 0 {
		fmt.Println("  No virtual accounts (set CREATE_VA=true to create one)")
	} else {
		for _, va := range vaList.VirtualAccounts {
			fmt.Printf("  %s — %s at %s (%s)\n", va.Provider, va.AccountNumber, va.BankName, va.Currency)
		}
	}

	// -----------------------------------------------------------------------
	// 8. List all wallets for this customer
	// -----------------------------------------------------------------------
	fmt.Println("\n→ Listing all wallets for customer...")
	walletList, err := client.Wallets.ListByCustomer(ctx, customerID)
	if err != nil {
		log.Fatalf("list customer wallets: %v", err)
	}
	fmt.Printf("  Found %d wallet(s)\n", walletList.Count)
	for _, w := range walletList.Wallets {
		fmt.Printf("    %s — %s %s\n", w.ID, w.Balance, w.Currency)
	}

	fmt.Println("\n✅ Wallet example complete!")
}
