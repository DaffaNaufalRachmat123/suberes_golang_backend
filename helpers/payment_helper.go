package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	httpClient *http.Client

	durianURL string
	xenditURL string
	flipURL   string

	durianKey string
	xenditKey string
	flipKey   string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		durianURL: os.Getenv("DURIANPAY_URL"),
		xenditURL: os.Getenv("XENDIT_PROD_URL"),
		flipURL:   os.Getenv("FLIP_DEV_URL"),
		durianKey: os.Getenv("DURIANPAY_API_KEY"),
		xenditKey: os.Getenv("XENDIT_API_KEY"),
		flipKey:   os.Getenv("FLIP_API_KEY"),
	}
}

func (c *Client) postJSON(ctx context.Context, endpoint string, payload interface{}, username string, headers map[string]string) ([]byte, error) {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(username, "")
	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) postForm(ctx context.Context, endpoint string, values url.Values, username string, headers map[string]string) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(username, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

//
// =============================
//  IMPLEMENTATION METHODS
// =============================
//

// 1. Create Virtual Account (Xendit)
func (c *Client) CreateVirtualAccount(ctx context.Context, body interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/callback_virtual_accounts", c.xenditURL)
	return c.postJSON(ctx, endpoint, body, c.xenditKey, nil)
}

// 2. Create Order Transaction (Durian)
func (c *Client) CreateOrderTransaction(ctx context.Context, payload interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/orders", c.durianURL)
	return c.postJSON(ctx, endpoint, payload, c.durianKey, nil)
}

// 3. Create Payment Charge (Durian)
func (c *Client) CreatePaymentCharge(ctx context.Context, payload interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/payments/charge", c.durianURL)
	return c.postJSON(ctx, endpoint, payload, c.durianKey, nil)
}

// 4. Create Bill Charge (Flip - form urlencoded)
func (c *Client) CreateBillCharge(ctx context.Context, body map[string]string) ([]byte, error) {
	values := url.Values{}
	for k, v := range body {
		values.Set(k, v)
	}

	endpoint := fmt.Sprintf("%s/v2/pwf/bill", c.flipURL)
	return c.postForm(ctx, endpoint, values, c.flipKey, nil)
}

// 5. Create Disbursement Flip (with idempotency key)
func (c *Client) CreateDisbursementFlip(ctx context.Context, idempotencyKey string, body map[string]string) ([]byte, error) {

	values := url.Values{}
	for k, v := range body {
		values.Set(k, v)
	}

	headers := map[string]string{
		"idempotency-key": idempotencyKey,
	}

	endpoint := fmt.Sprintf("%s/v3/disbursement", c.flipURL)
	return c.postForm(ctx, endpoint, values, c.flipKey, headers)
}

// 6. Create Ewallet Charge (Durian)
func (c *Client) CreateEwalletCharge(ctx context.Context, payload interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/payments/charge", c.durianURL)
	return c.postJSON(ctx, endpoint, payload, c.durianKey, nil)
}

// 7. Create Disbursement Charge Xendit
func (c *Client) CreateDisbursementChargeXendit(ctx context.Context, body interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/disbursements", c.xenditURL)
	return c.postJSON(ctx, endpoint, body, c.xenditKey, nil)
}

// 8. Create Ewallet Charge Xendit
func (c *Client) CreateEwalletChargeXendit(ctx context.Context, body interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/ewallets/charges", c.xenditURL)
	return c.postJSON(ctx, endpoint, body, c.xenditKey, nil)
}

// 9. VOID Charge Xendit (NO BODY, ONLY PATH ID)
func (c *Client) CreateVoidChargeXendit(ctx context.Context, id string) ([]byte, error) {

	endpoint := fmt.Sprintf("%s/ewallets/charges/%s/void", c.xenditURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.xenditKey, "")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// 10. Create Refund (Durian)
func (c *Client) CreateRefund(ctx context.Context, payload interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/refunds", c.durianURL)
	return c.postJSON(ctx, endpoint, payload, c.durianKey, nil)
}

// 11. Create Disbursement (Xendit + Idempotency Header JSON)
func (c *Client) CreateDisbursement(ctx context.Context, idempotencyKey string, payload interface{}) ([]byte, error) {

	headers := map[string]string{
		"X-IDEMPOTENCY-KEY": idempotencyKey,
	}

	endpoint := fmt.Sprintf("%s/disbursements", c.xenditURL)
	return c.postJSON(ctx, endpoint, payload, c.xenditKey, headers)
}

// 12. Validate Disbursement (Durian)
func (c *Client) ValidateDisbursement(ctx context.Context, payload interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/disbursement/validate", c.durianURL)
	return c.postJSON(ctx, endpoint, payload, c.durianKey, nil)
}
