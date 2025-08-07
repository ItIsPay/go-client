package itispay

import (
	"time"
)

// Invoice status constants
const (
	StatusNew         = "new"
	StatusPending     = "pending"
	StatusCompleted   = "completed"
	StatusExpired     = "expired"
	StatusCancelled   = "cancelled"
	StatusPaidPartial = "paid_partial"
)

// Sort order constants
const (
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

// Sort field constants
const (
	SortByCreatedAt    = "created_at"
	SortByUpdatedAt    = "updated_at"
	SortByFiatAmount   = "fiat_amount"
	SortByCryptoAmount = "crypto_amount"
)

// CreateInvoiceRequest represents the request to create a new invoice
type CreateInvoiceRequest struct {
	OrderID            string   `json:"order_id"`
	FiatAmount         *float64 `json:"fiat_amount,omitempty"`
	FiatCurrency       string   `json:"fiat_currency,omitempty"`
	CryptoAmount       *float64 `json:"crypto_amount,omitempty"`
	Currency           string   `json:"currency"`
	AllowedErrorPercent *int    `json:"allowed_error_percent,omitempty"`
	OrderName          string   `json:"order_name,omitempty"`
	ExpireMin          *int     `json:"expire_min,omitempty"`
	CallbackURL        string   `json:"callback_url,omitempty"`
}

// UpdateInvoiceRequest represents the request to update an invoice
type UpdateInvoiceRequest struct {
	Status string `json:"status"`
}

// ListInvoicesParams represents the parameters for listing invoices
type ListInvoicesParams struct {
	Page          int       `json:"page,omitempty"`
	PageSize      int       `json:"page_size,omitempty"`
	Status        string    `json:"status,omitempty"`
	Currency      string    `json:"currency,omitempty"`
	CreatedAfter  time.Time `json:"created_after,omitempty"`
	CreatedBefore time.Time `json:"created_before,omitempty"`
	SortBy        string    `json:"sort_by,omitempty"`
	SortOrder     string    `json:"sort_order,omitempty"`
}

// WebhookSimulateRequest represents the request to simulate a webhook
type WebhookSimulateRequest struct {
	InvoiceID string `json:"invoice_id"`
	Status    string `json:"status"`
}

// Invoice represents an invoice response
type Invoice struct {
	InvoiceID                    string            `json:"invoice_id"`
	UserID                       string            `json:"user_id"`
	ProjectID                    string            `json:"project_id"`
	OrderID                      string            `json:"order_id"`
	FiatAmount                   float64           `json:"fiat_amount"`
	FiatCurrency                 string            `json:"fiat_currency"`
	Currency                     string            `json:"currency"`
	CryptoAmount                 float64           `json:"crypto_amount"`
	CryptoAmountInUnits          *int64            `json:"crypto_amount_in_units,omitempty"`
	ActualCryptoAmountPaid       float64           `json:"actual_crypto_amount_paid"`
	ActualCryptoAmountPaidInUnits int64            `json:"actual_crypto_amount_paid_in_units"`
	AllowedErrorPercent          int               `json:"allowed_error_percent"`
	OrderName                    string            `json:"order_name"`
	ExpireMin                    int               `json:"expire_min"`
	CallbackURL                  string            `json:"callback_url"`
	Status                       string            `json:"status"`
	CreatedAt                    time.Time         `json:"created_at"`
	UpdatedAt                    time.Time         `json:"updated_at"`
	ExpiresAt                    time.Time         `json:"expires_at"`
	BlockchainDetails            *BlockchainDetails `json:"blockchain_details,omitempty"`
}

// BlockchainDetails represents blockchain information for an invoice
type BlockchainDetails struct {
	WalletID           string              `json:"walletId"`
	AccountID          string              `json:"accountId"`
	Currency           string              `json:"currency"`
	BlockchainAddress  string              `json:"blockchainAddress"`
	BlockchainNetwork  *BlockchainNetwork  `json:"blockchainNetwork,omitempty"`
	QRCode             string              `json:"qrcode,omitempty"`
}

// BlockchainNetwork represents blockchain network information
type BlockchainNetwork struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ListInvoicesResponse represents the response from listing invoices
type ListInvoicesResponse struct {
	Data       []Invoice      `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	TotalPages   int   `json:"total_pages"`
	TotalRecords int64 `json:"total_records"`
	HasNext      bool  `json:"has_next"`
	HasPrevious  bool  `json:"has_previous"`
}

// Currency represents a supported currency with detailed information
type Currency struct {
	CurrencyCode   string    `json:"currency_code"`
	IsCrypto       bool      `json:"is_crypto"`
	Precision      int       `json:"precision"`
	IsActive       bool      `json:"is_active"`
	WalletPattern  string    `json:"wallet_pattern,omitempty"`
	Network        string    `json:"network,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// CurrenciesResponse represents the response from getting supported currencies
type CurrenciesResponse struct {
	Currencies []Currency `json:"currencies"`
}

// RatesResponse represents the response from getting exchange rates
type RatesResponse struct {
	Rates map[string]float64 `json:"rates"`
}

// WebhookSimulateResponse represents the response from webhook simulation
type WebhookSimulateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// APIError represents an API error
type APIError struct {
	StatusCode int
	ErrorType  string
	Message    string
}

// Error returns the error message
func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.ErrorType
} 