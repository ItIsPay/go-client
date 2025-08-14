package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ItIsPay/go-client"
)

func main() {
	// Replace with your actual API key
	apiKey := "your-api-key-here"

	// Create a new client
	client := itispay.NewClient(apiKey)
	ctx := context.Background()

	fmt.Println("=== ItIsPay Go Client Example ===\n")

	// Example 1: Get supported currencies
	fmt.Println("1. Getting supported currencies...")
	currencies, err := client.GetCurrencies(ctx)
	if err != nil {
		log.Printf("Failed to get currencies: %v", err)
	} else {
		fmt.Println("Supported currencies:")
		for _, currency := range currencies.Currencies {
			if currency.IsCrypto {
				fmt.Printf("  %s (Crypto, %s network, precision: %d)\n",
					currency.CurrencyCode, currency.Network, currency.Precision)
			} else {
				fmt.Printf("  %s (Fiat, precision: %d)\n",
					currency.CurrencyCode, currency.Precision)
			}
		}
		fmt.Println()
	}

	// Example 2: Get current exchange rates
	fmt.Println("2. Getting current exchange rates...")
	rates, err := client.GetRates(ctx)
	if err != nil {
		log.Printf("Failed to get rates: %v", err)
	} else {
		fmt.Println("Current rates (EUR):")
		for currency, rate := range rates.Rates {
			fmt.Printf("  %s: %.2f\n", currency, rate)
		}
		fmt.Println()
	}

	// Example 3: Create an invoice
	fmt.Println("3. Creating a new invoice...")
	fiatAmount := 25.0
	expireMin := 30
	allowedError := 5

	invoice, err := client.CreateInvoice(ctx, itispay.CreateInvoiceRequest{
		OrderID:             fmt.Sprintf("ORDER-%d", time.Now().Unix()),
		FiatAmount:          &fiatAmount,
		FiatCurrency:        "EUR",
		Currency:            "BTC",
		AllowedErrorPercent: &allowedError,
		OrderName:           "Example Payment",
		ExpireMin:           &expireMin,
		CallbackURL:         "https://your-app.com/webhook",
	})
	if err != nil {
		log.Printf("Failed to create invoice: %v", err)
		return
	}

	fmt.Printf("✅ Invoice created successfully!\n")
	fmt.Printf("   Invoice ID: %s\n", invoice.InvoiceID)
	fmt.Printf("   Order ID: %s\n", invoice.OrderID)
	fmt.Printf("   Amount: %.2f %s = %f %s\n", invoice.FiatAmount, invoice.FiatCurrency, invoice.CryptoAmount, invoice.Currency)
	fmt.Printf("   Status: %s\n", invoice.Status)
	fmt.Printf("   Payment Address: %s\n", invoice.BlockchainDetails.BlockchainAddress)
	fmt.Printf("   Expires: %s\n", invoice.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Example 4: Get invoice details
	fmt.Println("4. Getting invoice details...")
	retrievedInvoice, err := client.GetInvoice(ctx, invoice.InvoiceID)
	if err != nil {
		log.Printf("Failed to get invoice: %v", err)
	} else {
		fmt.Printf("✅ Invoice retrieved: %s (Status: %s)\n\n", retrievedInvoice.InvoiceID, retrievedInvoice.Status)
	}

	// Example 5: List invoices
	fmt.Println("5. Listing recent invoices...")
	invoices, err := client.ListInvoices(ctx, itispay.ListInvoicesParams{
		Page:      1,
		PageSize:  5,
		SortBy:    itispay.SortByCreatedAt,
		SortOrder: itispay.SortOrderDesc,
	})
	if err != nil {
		log.Printf("Failed to list invoices: %v", err)
	} else {
		fmt.Printf("✅ Found %d invoices (Page %d of %d)\n",
			len(invoices.Items),
			invoices.Pagination.CurrentPage,
			invoices.Pagination.TotalPages,
		)
		for i, inv := range invoices.Items {
			fmt.Printf("   %d. %s - €%.2f %s (%s)\n",
				i+1,
				inv.OrderID,
				inv.FiatAmount,
				inv.FiatCurrency,
				inv.Status,
			)
		}
		fmt.Println()
	}

	// Example 6: Simulate a webhook (for testing)
	fmt.Println("6. Simulating webhook payment...")
	webhookResp, err := client.SimulateWebhook(ctx, invoice.InvoiceID, itispay.StatusCompleted)
	if err != nil {
		log.Printf("Failed to simulate webhook: %v", err)
	} else {
		fmt.Printf("✅ Webhook simulation: %s\n", webhookResp.Message)
	}

	// Example 7: Check updated invoice status
	fmt.Println("7. Checking updated invoice status...")
	time.Sleep(2 * time.Second) // Wait for webhook processing

	updatedInvoice, err := client.GetInvoice(ctx, invoice.InvoiceID)
	if err != nil {
		log.Printf("Failed to get updated invoice: %v", err)
	} else {
		fmt.Printf("✅ Invoice status updated: %s\n", updatedInvoice.Status)
		if updatedInvoice.Status == itispay.StatusCompleted {
			fmt.Printf("   Payment confirmed! Received %f BTC\n", updatedInvoice.ActualCryptoAmountPaid)
		}
	}

	fmt.Println("\n=== Example completed successfully! ===")
}

// webhookHandler demonstrates how to handle webhook callbacks from ItIsPay
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var webhook struct {
		InvoiceID                     string    `json:"invoice_id"`
		Status                        string    `json:"status"`
		OrderID                       string    `json:"order_id"`
		Currency                      string    `json:"currency"`
		CryptoAmount                  float64   `json:"crypto_amount"`
		FiatAmount                    float64   `json:"fiat_amount"`
		FiatCurrency                  string    `json:"fiat_currency"`
		ActualCryptoAmountPaid        float64   `json:"actual_crypto_amount_paid"`
		ActualCryptoAmountPaidInUnits int64     `json:"actual_crypto_amount_paid_in_units"`
		UpdatedAt                     time.Time `json:"updated_at"`
	}

	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Process the webhook based on status
	switch webhook.Status {
	case itispay.StatusCompleted:
		fmt.Printf("🎉 Payment completed for invoice %s (Order: %s)\n",
			webhook.InvoiceID, webhook.OrderID)
		fmt.Printf("   Amount: %f %s (€%.2f)\n",
			webhook.ActualCryptoAmountPaid, webhook.Currency, webhook.FiatAmount)
		// Here you would typically:
		// - Update your database
		// - Send confirmation email
		// - Fulfill the order
		// - etc.

	case itispay.StatusPaidPartial:
		fmt.Printf("⚠️  Partial payment received for invoice %s\n", webhook.InvoiceID)
		fmt.Printf("   Expected: %f %s, Received: %f %s\n",
			webhook.CryptoAmount, webhook.Currency,
			webhook.ActualCryptoAmountPaid, webhook.Currency)

	case itispay.StatusExpired:
		fmt.Printf("❌ Invoice %s expired\n", webhook.InvoiceID)

	case itispay.StatusCancelled:
		fmt.Printf("🚫 Invoice %s was cancelled\n", webhook.InvoiceID)

	default:
		fmt.Printf("ℹ️  Invoice %s status changed to: %s\n", webhook.InvoiceID, webhook.Status)
	}

	// Always respond with success to acknowledge receipt
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
