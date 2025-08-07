# ItIsPay Go Client

A clean, idiomatic Go client for the ItIsPay cryptocurrency payment gateway API. This package provides easy-to-use methods for creating and managing cryptocurrency invoices, handling payments, and integrating with the ItIsPay platform.

## Features

- ✅ **Zero Dependencies**: Uses only Go standard library
- ✅ **Clean API**: Idiomatic Go design with proper error handling
- ✅ **Context Support**: Full context.Context support for timeouts and cancellation
- ✅ **Type Safety**: Strongly typed requests and responses
- ✅ **Comprehensive**: Supports all ItIsPay API endpoints

## Installation

```bash
go get github.com/itispay/go-client
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/itispay/go-client"
)

func main() {
    // Create a new client
    client := itispay.NewClient("your-api-key")

    // Create an invoice
    ctx := context.Background()
    fiatAmount := 100.0
    
    invoice, err := client.CreateInvoice(ctx, itispay.CreateInvoiceRequest{
        OrderID:      "ORDER-12345",
        FiatAmount:   &fiatAmount,
        FiatCurrency: "EUR",
        Currency:     "BTC",
        OrderName:    "Premium Subscription",
        ExpireMin:    &[]int{30}[0],
        CallbackURL:  "https://your-app.com/webhook",
    })
    if err != nil {
        log.Fatal("Failed to create invoice:", err)
    }

    fmt.Printf("Created invoice: %s\n", invoice.InvoiceID)
    fmt.Printf("Payment address: %s\n", invoice.BlockchainDetails.BlockchainAddress)
    fmt.Printf("Amount: %f BTC\n", invoice.CryptoAmount)
}
```

## API Reference

### Client Creation

```go
// Create client with API credentials
client := itispay.NewClient("your-api-key")
```

**Note**: You can obtain your API key from your ItIsPay account dashboard after registration.

### Invoice Management

#### Create Invoice

```go
fiatAmount := 50.0
expireMin := 30

invoice, err := client.CreateInvoice(ctx, itispay.CreateInvoiceRequest{
    OrderID:            "ORDER-12345",
    FiatAmount:         &fiatAmount,
    FiatCurrency:       "EUR",
    Currency:           "BTC",
    AllowedErrorPercent: &[]int{5}[0],
    OrderName:          "Premium Subscription",
    ExpireMin:          &expireMin,
    CallbackURL:        "https://your-app.com/webhook",
})
```

#### Get Invoice

```go
invoice, err := client.GetInvoice(ctx, "invoice_7d4e8f2a-1b3c-4d5e-8f9a-2b3c4d5e6f7a")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Invoice status: %s\n", invoice.Status)
fmt.Printf("Amount paid: %f %s\n", invoice.ActualCryptoAmountPaid, invoice.Currency)
```

#### List Invoices

```go
// Basic listing
invoices, err := client.ListInvoices(ctx, itispay.ListInvoicesParams{})

// With filters and pagination
invoices, err := client.ListInvoices(ctx, itispay.ListInvoicesParams{
    Page:      1,
    PageSize:  20,
    Status:    itispay.StatusCompleted,
    Currency:  "BTC",
    SortBy:    itispay.SortByCreatedAt,
    SortOrder: itispay.SortOrderDesc,
})

for _, invoice := range invoices.Data {
    fmt.Printf("Invoice %s: %s - %f %s\n", 
        invoice.InvoiceID, 
        invoice.Status, 
        invoice.FiatAmount, 
        invoice.FiatCurrency,
    )
}
```

#### Update Invoice Status

```go
response, err := client.UpdateInvoiceStatus(ctx, "invoice_id", itispay.StatusCancelled)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Invoice %s updated to %s\n", response.InvoiceID, response.Status)
```

### Currency and Rates

#### Get Supported Currencies

```go
currencies, err := client.GetCurrencies(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Supported currencies:")
for _, currency := range currencies.Currencies {
    if currency.IsCrypto {
        fmt.Printf("- %s (Crypto, %s network)\n", currency.CurrencyCode, currency.Network)
    } else {
        fmt.Printf("- %s (Fiat)\n", currency.CurrencyCode)
    }
}
```

#### Get Exchange Rates

```go
rates, err := client.GetRates(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("BTC rate: $%.2f\n", rates.Rates["BTC"])
fmt.Printf("ETH rate: $%.2f\n", rates.Rates["ETH"])
```

### Webhook Testing

#### Simulate Webhook

```go
response, err := client.SimulateWebhook(ctx, "invoice_id", itispay.StatusCompleted)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Webhook simulation: %s\n", response.Message)
```

## Invoice Status Values

| Status | Description |
|--------|-------------|
| `itispay.StatusNew` | Invoice created, awaiting payment |
| `itispay.StatusPending` | Payment detected but not confirmed |
| `itispay.StatusCompleted` | Payment confirmed and within acceptable range |
| `itispay.StatusPaidPartial` | Payment received but below expected amount |
| `itispay.StatusExpired` | Invoice expired without payment |
| `itispay.StatusCancelled` | Invoice manually cancelled |

## Error Handling

The client returns typed errors for better error handling:

```go
invoice, err := client.GetInvoice(ctx, "invalid_id")
if err != nil {
    if apiErr, ok := err.(*itispay.APIError); ok {
        switch apiErr.StatusCode {
        case 401:
            fmt.Println("Authentication failed")
        case 404:
            fmt.Println("Invoice not found")
        case 400:
            fmt.Printf("Bad request: %s\n", apiErr.Message)
        default:
            fmt.Printf("API error: %s\n", apiErr.Message)
        }
    } else {
        fmt.Printf("Network error: %v\n", err)
    }
    return
}
```

## Webhook Integration

To handle webhook callbacks from ItIsPay, create an HTTP handler:

```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var webhook struct {
        InvoiceID                    string    `json:"invoice_id"`
        Status                       string    `json:"status"`
        OrderID                      string    `json:"order_id"`
        Currency                     string    `json:"currency"`
        CryptoAmount                 float64   `json:"crypto_amount"`
        FiatAmount                   float64   `json:"fiat_amount"`
        FiatCurrency                 string    `json:"fiat_currency"`
        ActualCryptoAmountPaid       float64   `json:"actual_crypto_amount_paid"`
        ActualCryptoAmountPaidInUnits int64    `json:"actual_crypto_amount_paid_in_units"`
        UpdatedAt                    time.Time `json:"updated_at"`
    }

    if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Process the webhook
    switch webhook.Status {
    case itispay.StatusCompleted:
        fmt.Printf("Payment completed for invoice %s\n", webhook.InvoiceID)
        // Update your database, send confirmation email, etc.
    case itispay.StatusPaidPartial:
        fmt.Printf("Partial payment received for invoice %s\n", webhook.InvoiceID)
    case itispay.StatusExpired:
        fmt.Printf("Invoice %s expired\n", webhook.InvoiceID)
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}
```

## Complete Example

Here's a complete example showing a typical payment flow:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/itispay/go-client"
)

func main() {
    client := itispay.NewClient("your-api-key")
    ctx := context.Background()

    // 1. Create an invoice
    fiatAmount := 25.0
    expireMin := 30
    
    invoice, err := client.CreateInvoice(ctx, itispay.CreateInvoiceRequest{
        OrderID:      fmt.Sprintf("ORDER-%d", time.Now().Unix()),
        FiatAmount:   &fiatAmount,
        FiatCurrency: "USD",
        Currency:     "BTC",
        OrderName:    "Test Payment",
        ExpireMin:    &expireMin,
        CallbackURL:  "https://your-app.com/webhook",
    })
    if err != nil {
        log.Fatal("Failed to create invoice:", err)
    }

    fmt.Printf("Created invoice: %s\n", invoice.InvoiceID)
    fmt.Printf("Pay %f BTC to: %s\n", 
        invoice.CryptoAmount, 
        invoice.BlockchainDetails.BlockchainAddress,
    )

    // 2. Simulate a payment (for testing)
    _, err = client.SimulateWebhook(ctx, invoice.InvoiceID, itispay.StatusCompleted)
    if err != nil {
        log.Printf("Failed to simulate payment: %v", err)
    }

    // 3. Check invoice status
    time.Sleep(2 * time.Second) // Wait for webhook processing
    
    updatedInvoice, err := client.GetInvoice(ctx, invoice.InvoiceID)
    if err != nil {
        log.Fatal("Failed to get invoice:", err)
    }

    fmt.Printf("Invoice status: %s\n", updatedInvoice.Status)
    if updatedInvoice.Status == itispay.StatusCompleted {
        fmt.Printf("Payment confirmed! Received %f BTC\n", 
            updatedInvoice.ActualCryptoAmountPaid,
        )
    }
}
```

## Development Setup

This project uses Go workspaces for local development. The `go.work` file enables working with multiple modules simultaneously.

### Local Development

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd itispay-go-client
   ```

2. **Run the example:**
   ```bash
   cd example
   go run main.go
   ```

The workspace automatically handles the module dependencies, so you don't need to use `go get` or `replace` directives during development.

### Production Usage

```bash
go get github.com/itispay/go-client
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For API support and integration assistance:

- **Documentation**: Review the [ItIsPay API Reference](https://api-docs.itispay.com/invoice)
- **Testing**: Use the webhook simulation endpoint for integration testing
- **Issues**: Report bugs and feature requests via GitHub issues