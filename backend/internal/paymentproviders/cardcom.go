package paymentproviders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CardcomAdapter implements Cardcom API v11. Cardcom's LowProfile webhook is
// treated as a notification only; callers must invoke VerifyPayment before
// changing the local payment or sale to paid.
type CardcomAdapter struct {
	config Config
	client *http.Client
}

func NewCardcomAdapter(config Config) *CardcomAdapter {
	if config.BaseURL == "" {
		config.BaseURL = "https://secure.cardcom.solutions/api/v11"
	}
	if config.HTTPTimeout <= 0 {
		config.HTTPTimeout = 20 * time.Second
	}
	return &CardcomAdapter{config: config, client: &http.Client{Timeout: config.HTTPTimeout}}
}

func (a *CardcomAdapter) Name() ProviderName { return ProviderCardcom }

func (a *CardcomAdapter) TestConnection(ctx context.Context) error {
	terminal, err := strconv.Atoi(a.config.TerminalID)
	if err != nil || terminal <= 0 || strings.TrimSpace(a.config.APIKey) == "" || strings.TrimSpace(a.config.APISecret) == "" {
		return ErrNotConfigured
	}
	_, err = a.post(ctx, "/Transactions/ListTransactions", map[string]any{
		"ApiName":          a.config.APIKey,
		"ApiPassword":      a.config.APISecret,
		"FromDate":         time.Now().Format("02012006"),
		"ToDate":           time.Now().Format("02012006"),
		"TranStatus":       "Success",
		"Page":             1,
		"Page_size":        1,
		"LimitForTerminal": terminal,
	})
	return err
}

func (a *CardcomAdapter) CreatePayment(ctx context.Context, request PaymentRequest) (PaymentResult, error) {
	terminal, err := strconv.Atoi(a.config.TerminalID)
	if err != nil || terminal <= 0 {
		return PaymentResult{}, ErrNotConfigured
	}
	coinID, err := cardcomCoinID(request.Currency)
	if err != nil {
		return PaymentResult{}, err
	}
	payload := map[string]any{
		"TerminalNumber":     terminal,
		"ApiName":            a.config.APIKey,
		"Amount":             float64(request.Amount) / 100,
		"ISOCoinId":          coinID,
		"Operation":          "ChargeOnly",
		"ReturnValue":        request.OrderID.String(),
		"SuccessRedirectUrl": request.SuccessURL,
		"FailedRedirectUrl":  request.FailureURL,
		"CancelRedirectUrl":  request.CancelURL,
		"WebHookUrl":         a.config.WebhookURL,
		"ProductName":        request.Description,
	}
	if request.IdempotencyKey != "" {
		payload["ExternalUniqTranId"] = request.IdempotencyKey
	}
	response, err := a.post(ctx, "/LowProfile/Create", payload)
	if err != nil {
		return PaymentResult{}, err
	}
	var result struct {
		ResponseCode int    `json:"ResponseCode"`
		Description  string `json:"Description"`
		LowProfileID string `json:"LowProfileId"`
		URL          string `json:"Url"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return PaymentResult{}, fmt.Errorf("decode Cardcom create response: %w", err)
	}
	if result.ResponseCode != 0 || result.LowProfileID == "" {
		return PaymentResult{}, fmt.Errorf("Cardcom create payment failed (%d): %s", result.ResponseCode, result.Description)
	}
	return PaymentResult{ProviderPaymentID: result.LowProfileID, Status: StatusPending, Amount: request.Amount, Currency: request.Currency, CheckoutURL: result.URL, Raw: response}, nil
}

func (a *CardcomAdapter) GetPaymentStatus(ctx context.Context, request StatusRequest) (PaymentResult, error) {
	return a.verify(ctx, request)
}

func (a *CardcomAdapter) VerifyPayment(ctx context.Context, request StatusRequest) (PaymentResult, error) {
	return a.verify(ctx, request)
}

func (a *CardcomAdapter) verify(ctx context.Context, request StatusRequest) (PaymentResult, error) {
	if request.ProviderPaymentID == "" {
		return PaymentResult{}, ErrInvalidProviderData
	}
	response, err := a.post(ctx, "/LowProfile/GetLpResult", map[string]any{
		"TerminalNumber": atoiOrZero(a.config.TerminalID),
		"ApiName":        a.config.APIKey,
		"ApiPassword":    a.config.APISecret,
		"LowProfileId":   request.ProviderPaymentID,
	})
	if err != nil {
		return PaymentResult{}, err
	}
	var result struct {
		ResponseCode    int    `json:"ResponseCode"`
		Description     string `json:"Description"`
		LowProfileID    string `json:"LowProfileId"`
		TranzactionInfo *struct {
			ResponseCode  int     `json:"ResponseCode"`
			Amount        float64 `json:"Amount"`
			CoinID        int     `json:"CoinId"`
			TranzactionID int64   `json:"TranzactionId"`
		} `json:"TranzactionInfo"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return PaymentResult{}, fmt.Errorf("decode Cardcom status response: %w", err)
	}
	status := StatusProcessing
	amount := int64(0)
	if result.ResponseCode != 0 || result.TranzactionInfo == nil || result.TranzactionInfo.ResponseCode != 0 {
		status = StatusFailed
	} else {
		status = StatusPaid
		amount = int64(result.TranzactionInfo.Amount * 100)
	}
	transactionID := ""
	if result.TranzactionInfo != nil {
		transactionID = strconv.FormatInt(result.TranzactionInfo.TranzactionID, 10)
	}
	return PaymentResult{ProviderPaymentID: result.LowProfileID, ProviderTransactionID: transactionID, Status: status, Amount: amount, Raw: response}, nil
}

func (a *CardcomAdapter) CancelPayment(ctx context.Context, request StatusRequest) (PaymentResult, error) {
	return a.refund(ctx, request.ProviderTransactionID, 0, true, "cancel")
}

func (a *CardcomAdapter) RefundPayment(ctx context.Context, request RefundRequest) (RefundResult, error) {
	result, err := a.refund(ctx, request.ProviderTransactionID, request.Amount, false, request.Reason)
	if err != nil {
		return RefundResult{}, err
	}
	return RefundResult{ProviderRefundID: result.ProviderPaymentID, Status: result.Status, Amount: request.Amount, Raw: result.Raw}, nil
}

func (a *CardcomAdapter) refund(ctx context.Context, providerPaymentID string, amount int64, cancelOnly bool, reason string) (PaymentResult, error) {
	transactionID, err := strconv.ParseInt(providerPaymentID, 10, 64)
	if err != nil {
		return PaymentResult{}, fmt.Errorf("Cardcom refund requires transaction id: %w", err)
	}
	payload := map[string]any{
		"ApiName":              a.config.APIKey,
		"ApiPassword":          a.config.APISecret,
		"TransactionId":        transactionID,
		"CancelOnly":           cancelOnly,
		"ExternalRefundDealId": reason,
	}
	if amount > 0 {
		payload["PartialSum"] = float64(amount) / 100
	}
	response, err := a.post(ctx, "/Transactions/RefundByTransactionId", payload)
	if err != nil {
		return PaymentResult{}, err
	}
	var result struct {
		ResponseCode int    `json:"ResponseCode"`
		Description  string `json:"Description"`
		NewID        int64  `json:"NewTranzactionId"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return PaymentResult{}, fmt.Errorf("decode Cardcom refund response: %w", err)
	}
	if result.ResponseCode != 0 {
		return PaymentResult{}, fmt.Errorf("Cardcom refund failed (%d): %s", result.ResponseCode, result.Description)
	}
	status := StatusRefunded
	if amount > 0 {
		status = StatusPartiallyRefunded
	}
	return PaymentResult{ProviderPaymentID: strconv.FormatInt(result.NewID, 10), Status: status, Amount: amount, Raw: response}, nil
}

func (a *CardcomAdapter) HandleWebhook(_ context.Context, _ map[string]string, body []byte) (WebhookEvent, error) {
	var event struct {
		LowProfileID    string `json:"LowProfileId"`
		ReturnValue     string `json:"ReturnValue"`
		ResponseCode    int    `json:"ResponseCode"`
		TranzactionInfo *struct {
			Amount float64 `json:"Amount"`
			CoinID int     `json:"CoinId"`
		} `json:"TranzactionInfo"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		return WebhookEvent{}, fmt.Errorf("decode Cardcom webhook: %w", err)
	}
	if event.LowProfileID == "" {
		return WebhookEvent{}, ErrInvalidProviderData
	}
	status := StatusProcessing
	if event.ResponseCode != 0 {
		status = StatusFailed
	}
	amount := int64(0)
	if event.TranzactionInfo != nil {
		amount = int64(event.TranzactionInfo.Amount * 100)
	}
	return WebhookEvent{EventID: event.LowProfileID, ProviderPaymentID: event.LowProfileID, Status: status, Amount: amount, Raw: append([]byte(nil), body...)}, nil
}

func (a *CardcomAdapter) post(ctx context.Context, path string, payload map[string]any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(a.config.BaseURL, "/")+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Cardcom request: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Cardcom HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return responseBody, nil
}

func cardcomCoinID(currency string) (int, error) {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "ILS":
		return 1, nil
	case "USD":
		return 2, nil
	default:
		return 0, fmt.Errorf("Cardcom does not support configured currency %q in this adapter", currency)
	}
}

func atoiOrZero(value string) int {
	result, _ := strconv.Atoi(value)
	return result
}
