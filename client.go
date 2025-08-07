package itispay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultBaseURL is the default ItIsPay API base URL
	DefaultBaseURL = "https://api.itispay.com/api/v1"
	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second
)

// Client represents an ItIsPay API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new ItIsPay API client
func NewClient(apiKey string) *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// doRequest performs an HTTP request and unmarshals the response
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		var apiError ErrorResponse
		if err := json.Unmarshal(respBody, &apiError); err != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			ErrorType:  apiError.Error,
			Message:    apiError.Message,
		}
	}

	return respBody, nil
}

// CreateInvoice creates a new cryptocurrency invoice
func (c *Client) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
	// Debug: print the request being sent
	reqJSON, _ := json.MarshalIndent(req, "", "  ")
	fmt.Printf("DEBUG: Creating invoice with request:\n%s\n", string(reqJSON))
	
	respBody, err := c.doRequest(ctx, "POST", "/invoices", req)
	if err != nil {
		return nil, err
	}

	var invoice Invoice
	if err := json.Unmarshal(respBody, &invoice); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice response: %w", err)
	}

	return &invoice, nil
}

// GetInvoice retrieves a specific invoice by ID
func (c *Client) GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error) {
	respBody, err := c.doRequest(ctx, "GET", "/invoices/"+invoiceID, nil)
	if err != nil {
		return nil, err
	}

	var invoice Invoice
	if err := json.Unmarshal(respBody, &invoice); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice response: %w", err)
	}

	return &invoice, nil
}

// ListInvoices retrieves a paginated list of invoices with optional filtering
func (c *Client) ListInvoices(ctx context.Context, params ListInvoicesParams) (*ListInvoicesResponse, error) {
	// Build query parameters
	queryParams := url.Values{}
	if params.Page > 0 {
		queryParams.Set("page", fmt.Sprintf("%d", params.Page))
	}
	if params.PageSize > 0 {
		queryParams.Set("page_size", fmt.Sprintf("%d", params.PageSize))
	}
	if params.Status != "" {
		queryParams.Set("status", params.Status)
	}
	if params.Currency != "" {
		queryParams.Set("currency", params.Currency)
	}
	if !params.CreatedAfter.IsZero() {
		queryParams.Set("created_after", params.CreatedAfter.Format(time.RFC3339))
	}
	if !params.CreatedBefore.IsZero() {
		queryParams.Set("created_before", params.CreatedBefore.Format(time.RFC3339))
	}
	if params.SortBy != "" {
		queryParams.Set("sort_by", params.SortBy)
	}
	if params.SortOrder != "" {
		queryParams.Set("sort_order", params.SortOrder)
	}

	path := "/invoices"
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	respBody, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response ListInvoicesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoices response: %w", err)
	}

	return &response, nil
}

// GetCurrencies retrieves the list of supported currencies
func (c *Client) GetCurrencies(ctx context.Context) (*CurrenciesResponse, error) {
	respBody, err := c.doRequest(ctx, "GET", "/currencies", nil)
	if err != nil {
		return nil, err
	}

	// The API returns an array of currency objects directly
	var currencies []Currency
	if err := json.Unmarshal(respBody, &currencies); err != nil {
		return nil, fmt.Errorf("failed to unmarshal currencies response: %w", err)
	}

	return &CurrenciesResponse{Currencies: currencies}, nil
}

// GetRates retrieves current exchange rates for supported cryptocurrencies
func (c *Client) GetRates(ctx context.Context) (*RatesResponse, error) {
	respBody, err := c.doRequest(ctx, "GET", "/rates", nil)
	if err != nil {
		return nil, err
	}

	var response RatesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rates response: %w", err)
	}

	return &response, nil
}

// UpdateInvoiceStatus updates the status of an existing invoice
func (c *Client) UpdateInvoiceStatus(ctx context.Context, invoiceID string, status string) (*Invoice, error) {
	req := UpdateInvoiceRequest{Status: status}
	respBody, err := c.doRequest(ctx, "PATCH", "/invoices/"+invoiceID, req)
	if err != nil {
		return nil, err
	}

	var invoice Invoice
	if err := json.Unmarshal(respBody, &invoice); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice response: %w", err)
	}

	return &invoice, nil
}

// SimulateWebhook simulates a webhook callback for testing purposes (no authentication required)
func (c *Client) SimulateWebhook(ctx context.Context, invoiceID, status string) (*WebhookSimulateResponse, error) {
	req := WebhookSimulateRequest{
		InvoiceID: invoiceID,
		Status:    status,
	}
	respBody, err := c.doRequest(ctx, "POST", "/webhooks/simulate", req)
	if err != nil {
		return nil, err
	}

	var response WebhookSimulateResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook simulate response: %w", err)
	}

	return &response, nil
}
