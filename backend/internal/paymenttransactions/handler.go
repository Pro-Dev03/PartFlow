package paymenttransactions

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/paymentproviders"
	"github.com/partflow/smart-store/internal/secrets"
	partflowdb "github.com/partflow/smart-store/pkg/database"
	"github.com/partflow/smart-store/pkg/middleware"
)

type Handler struct {
	db *sqlx.DB
}

func NewHandler(db *sqlx.DB) *Handler { return &Handler{db: db} }

type createRequest struct {
	OrderID        uuid.UUID  `json:"order_id"`
	SaleID         *uuid.UUID `json:"sale_id,omitempty"`
	AmountMinor    int64      `json:"amount_minor" binding:"required,gt=0"`
	Currency       string     `json:"currency"`
	Description    string     `json:"description"`
	IdempotencyKey string     `json:"idempotency_key" binding:"required"`
	SuccessURL     string     `json:"success_url"`
	FailureURL     string     `json:"failure_url"`
	CancelURL      string     `json:"cancel_url"`
}

type refundRequest struct {
	AmountMinor    int64  `json:"amount_minor" binding:"required,gt=0"`
	Currency       string `json:"currency"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
	Reason         string `json:"reason"`
}

func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup, protected *gin.RouterGroup) {
	v1.POST("/payment-webhooks/:provider", h.Webhook)
	transactions := protected.Group("/payment-transactions")
	transactions.POST("", h.Create)
	transactions.GET("/:id", h.Get)
	transactions.POST("/:id/verify", h.Verify)
	transactions.POST("/:id/cancel", h.Cancel)
	transactions.POST("/:id/refund", h.Refund)
	protected.POST("/payment-providers/test-connection", h.TestConnection)
	protected.GET("/payment-providers", h.ListProviders)
}

func (h *Handler) Create(c *gin.Context) {
	var request createRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	providerName, service, err := h.service(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	currency := request.Currency
	if request.OrderID == uuid.Nil {
		request.OrderID = uuid.New()
	}
	if currency == "" {
		currency = h.setting("currency", "ILS")
	}
	transaction, err := service.CreatePayment(c.Request.Context(), paymentproviders.PaymentRequest{
		OrderID: request.OrderID, SaleID: request.SaleID, Amount: request.AmountMinor, Currency: currency,
		Description: request.Description, IdempotencyKey: request.IdempotencyKey,
		SuccessURL: request.SuccessURL, FailureURL: request.FailureURL, CancelURL: request.CancelURL,
	}, providerName)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": transaction})
}

func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment transaction id"})
		return
	}
	transaction, err := NewRepository(h.db).GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment transaction not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": transaction})
}

func (h *Handler) Verify(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment transaction id"})
		return
	}
	_, service, err := h.service(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	transaction, err := service.VerifyPayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": transaction})
}

func (h *Handler) Refund(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment transaction id"})
		return
	}
	var request refundRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, service, err := h.service(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	refund, transaction, err := service.RefundPayment(c.Request.Context(), id, paymentproviders.RefundRequest{Amount: request.AmountMinor, Currency: request.Currency, IdempotencyKey: request.IdempotencyKey, Reason: request.Reason}, userID(c))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"refund": refund, "transaction": transaction}})
}

func (h *Handler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment transaction id"})
		return
	}
	_, service, err := h.service(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	transaction, err := service.CancelPayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": transaction})
}

func (h *Handler) TestConnection(c *gin.Context) {
	providerName, adapter, err := h.adapterFromSettings()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := adapter.TestConnection(c.Request.Context()); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "provider": providerName, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "provider": providerName, "status": "connected"})
}

func (h *Handler) ListProviders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": paymentproviders.NewRegistry().Descriptors()})
}

func (h *Handler) Webhook(c *gin.Context) {
	// A provider callback has no authenticated account context. Until callbacks
	// carry a verifiable per-tenant route and scope their database context, do
	// not let an unscoped request inherit another request's session tenant.
	usesCloudDatabase := h.db != nil && !strings.EqualFold(h.db.DriverName(), "sqlite")
	if partflowdb.TenantIsolationEnabled() || usesCloudDatabase {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "payment webhook routing must be configured for tenant isolation",
			"code":  "PAYMENT_WEBHOOK_TENANT_ROUTING_REQUIRED",
		})
		return
	}
	providerName := paymentproviders.ProviderName(strings.ToLower(strings.TrimSpace(c.Param("provider"))))
	adapter, err := h.adapter(providerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read webhook body"})
		return
	}
	headers := map[string]string{"content-type": c.GetHeader("Content-Type"), "signature": c.GetHeader("Stripe-Signature")}
	event, err := adapter.HandleWebhook(c.Request.Context(), headers, body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	repo := NewRepository(h.db)
	transaction, lookupErr := repo.GetByProviderPaymentID(c.Request.Context(), string(providerName), event.ProviderPaymentID)
	var transactionID *uuid.UUID
	if lookupErr == nil {
		transactionID = &transaction.ID
	}
	if err := repo.RecordWebhookEvent(c.Request.Context(), &WebhookEvent{Provider: string(providerName), ProviderEventID: event.EventID, PaymentTransactionID: transactionID, Payload: string(body)}); errors.Is(err, ErrDuplicateWebhook) {
		c.JSON(http.StatusAccepted, gin.H{"success": true, "status": "duplicate"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record webhook event"})
		return
	}
	if lookupErr != nil {
		c.JSON(http.StatusAccepted, gin.H{"success": true, "status": "received"})
		return
	}
	_, service, err := h.service(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// The webhook is only a trigger. Verification performs the authoritative
	// provider lookup before local state can become paid.
	if _, err := service.VerifyPayment(c.Request.Context(), transaction.ID); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "status": "verified"})
}

func (h *Handler) service(c *gin.Context) (paymentproviders.ProviderName, *Service, error) {
	name, adapter, err := h.adapterFromSettings()
	if err != nil {
		return name, nil, err
	}
	service := NewService(NewRepository(h.db))
	service.RegisterProvider(adapter)
	return name, service, nil
}

func (h *Handler) adapterFromSettings() (paymentproviders.ProviderName, paymentproviders.PaymentProvider, error) {
	name := paymentproviders.ProviderName(strings.ToLower(h.setting("payment_provider", "manual")))
	adapter, err := h.adapter(name)
	return name, adapter, err
}

func (h *Handler) adapter(name paymentproviders.ProviderName) (paymentproviders.PaymentProvider, error) {
	config := paymentproviders.Config{
		Provider:      name,
		Environment:   h.setting("payment_environment", "test"),
		Currency:      h.setting("currency", "ILS"),
		MerchantID:    h.setting("payment_merchant_id", ""),
		TerminalID:    h.setting("payment_terminal_id", ""),
		APIKey:        h.setting("payment_public_key", ""),
		APISecret:     h.secretSetting("payment_secret_key"),
		WebhookURL:    h.setting("payment_webhook_url", ""),
		WebhookSecret: h.secretSetting("payment_webhook_secret"),
	}
	return paymentproviders.NewRegistry().Build(name, config)
}

func (h *Handler) setting(key, fallback string) string {
	var value string
	if err := h.db.Get(&value, `SELECT value FROM settings WHERE key = $1`, key); err != nil || strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func (h *Handler) secretSetting(key string) string {
	value := h.setting(key, "")
	decrypted, err := secrets.Decrypt(value)
	if err != nil {
		return ""
	}
	return decrypted
}

func userID(c *gin.Context) *uuid.UUID {
	id := middleware.GetUserID(c)
	if id == uuid.Nil {
		return nil
	}
	return &id
}
