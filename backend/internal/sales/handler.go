package sales

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/errors"
	"github.com/partflow/smart-store/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateSale creates a new sale
// @Summary Create Sale
// @Description Create a new sale with profit calculation
// @Tags sales
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateSaleRequest true "Sale data"
// @Success 201 {object} response.Response{data=SaleWithItems}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/sales [post]
func (h *Handler) CreateSale(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		errors.HandleError(c, errors.NewValidationError("organization_id required", nil))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errors.HandleError(c, errors.NewValidationError("user_id required", nil))
		return
	}

	var req CreateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.ValidateRequest(err))
		return
	}

	sale, err := h.service.CreateSale(c.Request.Context(), organizationID.(uuid.UUID), userID.(uuid.UUID), &req)
	if err != nil {
		switch err {
		case ErrInsufficientStock:
			errors.HandleError(c, errors.NewBusinessError("Insufficient stock for one or more products", err))
		case ErrInvalidCustomer:
			errors.HandleError(c, errors.NewValidationError("Invalid customer", err))
		case ErrInvalidPaymentMethod:
			errors.HandleError(c, errors.NewValidationError("Invalid payment method", err))
		default:
			errors.HandleError(c, errors.WrapError(err, "Failed to create sale"))
		}
		return
	}

	// Get sale with items and profit
	saleWithItems, err := h.service.GetSale(c.Request.Context(), sale.ID)
	if err != nil {
		errors.HandleError(c, errors.WrapError(err, "Failed to retrieve sale details"))
		return
	}

	response.Created(c, saleWithItems, "Sale created successfully")
}

// GetSale retrieves a sale by ID
// @Summary Get Sale
// @Description Get a sale by ID with items and profit
// @Tags sales
// @Produce json
// @Security Bearer
// @Param id path string true "Sale ID"
// @Success 200 {object} response.Response{data=SaleWithItems}
// @Failure 404 {object} response.Response
// @Router /api/v1/sales/{id} [get]
func (h *Handler) GetSale(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("invalid sale id", err))
		return
	}

	sale, err := h.service.GetSale(c.Request.Context(), id)
	if err != nil {
		if err == ErrSaleNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Sale", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to retrieve sale"))
		return
	}

	response.OK(c, sale, "Sale retrieved successfully")
}

// ListSales retrieves sales with pagination and filters
// @Summary List Sales
// @Description Get sales with pagination and filters
// @Tags sales
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Param customer_id query string false "Filter by customer ID"
// @Param start_date query string false "Filter by start date"
// @Param end_date query string false "Filter by end date"
// @Success 200 {object} response.Response{data=[]Sale}
// @Router /api/v1/sales [get]
func (h *Handler) ListSales(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		errors.HandleError(c, errors.NewValidationError("organization_id required", nil))
		return
	}

	req := &SalesListRequest{
		Page:    1,
		PerPage: 20,
	}

	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}

	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}

	req.Status = c.Query("status")
	req.CustomerID = c.Query("customer_id")
	req.StartDate = c.Query("start_date")
	req.EndDate = c.Query("end_date")

	filters := make(map[string]interface{})
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.CustomerID != "" {
		if id, err := uuid.Parse(req.CustomerID); err == nil {
			filters["customer_id"] = id
		}
	}
	if req.StartDate != "" {
		filters["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		filters["end_date"] = req.EndDate
	}

	sales, total, err := h.service.ListSales(c.Request.Context(), organizationID.(uuid.UUID), req.Page, req.PerPage, filters)
	if err != nil {
		errors.HandleError(c, errors.WrapError(err, "Failed to retrieve sales"))
		return
	}

	response.OK(c, map[string]interface{}{
		"sales":     sales,
		"total":     total,
		"page":      req.Page,
		"per_page":  req.PerPage,
	}, "Sales retrieved successfully")
}

// UpdateSalePayment updates the payment for a sale
// @Summary Update Sale Payment
// @Description Update payment information for a sale
// @Tags sales
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Sale ID"
// @Param request body UpdatePaymentRequest true "Payment data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/sales/{id}/payment [post]
func (h *Handler) UpdateSalePayment(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		errors.HandleError(c, errors.NewValidationError("organization_id required", nil))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errors.HandleError(c, errors.NewValidationError("user_id required", nil))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("invalid sale id", err))
		return
	}

	var req UpdatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.ValidateRequest(err))
		return
	}

	if err := h.service.UpdateSalePayment(c.Request.Context(), organizationID.(uuid.UUID), userID.(uuid.UUID), id, req.Amount, req.PaymentMethod); err != nil {
		switch err {
		case ErrSaleNotFound:
			errors.HandleError(c, errors.NewNotFoundError("Sale", err))
		case ErrInvalidPayment:
			errors.HandleError(c, errors.NewValidationError("invalid payment amount", err))
		default:
			errors.HandleError(c, errors.WrapError(err, "Failed to update payment"))
		}
		return
	}

	response.OK(c, map[string]interface{}{"message": "payment updated successfully"}, "Payment updated successfully")
}

// CancelSale cancels a sale
// @Summary Cancel Sale
// @Description Cancel a sale
// @Tags sales
// @Security Bearer
// @Param id path string true "Sale ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/sales/{id}/cancel [post]
func (h *Handler) CancelSale(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("invalid sale id", err))
		return
	}

	if err := h.service.CancelSale(c.Request.Context(), id); err != nil {
		switch err {
		case ErrSaleNotFound:
			errors.HandleError(c, errors.NewNotFoundError("Sale", err))
		case ErrInvalidSaleStatus:
			errors.HandleError(c, errors.NewBusinessError("sale cannot be cancelled", err))
		default:
			errors.HandleError(c, errors.WrapError(err, "Failed to cancel sale"))
		}
		return
	}

	response.OK(c, map[string]interface{}{"message": "sale cancelled successfully"}, "Sale cancelled successfully")
}

// GetSalesSummary retrieves sales summary for a period
// @Summary Get Sales Summary
// @Description Get sales summary statistics for a period
// @Tags sales
// @Produce json
// @Security Bearer
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=SalesSummary}
// @Router /api/v1/sales/summary [get]
func (h *Handler) GetSalesSummary(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		response.BadRequest(c, "start_date and end_date are required")
		return
	}

	summary, err := h.service.GetSalesSummary(c.Request.Context(), organizationID.(uuid.UUID), startDate, endDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, summary, "Sales summary retrieved successfully")
}

// GetTopSellingProducts retrieves top selling products
// @Summary Get Top Selling Products
// @Description Get top selling products
// @Tags sales
// @Produce json
// @Security Bearer
// @Param limit query int false "Result limit" default(10)
// @Success 200 {object} response.Response{data=[]TopSellingProduct}
// @Router /api/v1/sales/top-products [get]
func (h *Handler) GetTopSellingProducts(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	products, err := h.service.GetTopSellingProducts(c.Request.Context(), organizationID.(uuid.UUID), limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, products, "Top selling products retrieved successfully")
}

// CreateTransaction creates a new financial transaction
// @Summary Create Transaction
// @Description Create a new financial transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body Transaction true "Transaction data"
// @Success 201 {object} response.Response{data=Transaction}
// @Failure 400 {object} response.Response
// @Router /api/v1/transactions [post]
func (h *Handler) CreateTransaction(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	var tx Transaction
	if err := c.ShouldBindJSON(&tx); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.CreateTransaction(c.Request.Context(), organizationID.(uuid.UUID), &tx); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, tx, "Transaction created successfully")
}

// GetTransaction retrieves a transaction by ID
// @Summary Get Transaction
// @Description Get a transaction by ID
// @Tags transactions
// @Produce json
// @Security Bearer
// @Param id path string true "Transaction ID"
// @Success 200 {object} response.Response{data=Transaction}
// @Failure 404 {object} response.Response
// @Router /api/v1/transactions/{id} [get]
func (h *Handler) GetTransaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid transaction id")
		return
	}

	tx, err := h.service.GetTransaction(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "transaction not found")
		return
	}

	response.OK(c, tx, "Transaction retrieved successfully")
}

// ListTransactions retrieves transactions with pagination and filters
// @Summary List Transactions
// @Description Get transactions with pagination and filters
// @Tags transactions
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param type query string false "Filter by type"
// @Param status query string false "Filter by status"
// @Param start_date query string false "Filter by start date"
// @Param end_date query string false "Filter by end date"
// @Success 200 {object} response.Response{data=[]Transaction}
// @Router /api/v1/transactions [get]
func (h *Handler) ListTransactions(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	page := 1
	perPage := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	filters := make(map[string]interface{})
	if txType := c.Query("type"); txType != "" {
		filters["type"] = txType
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if startDate := c.Query("start_date"); startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		filters["end_date"] = endDate
	}

	transactions, total, err := h.service.ListTransactions(c.Request.Context(), organizationID.(uuid.UUID), page, perPage, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, map[string]interface{}{
		"transactions": transactions,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
	}, "Transactions retrieved successfully")
}

// CalculateProfitForPeriod calculates profit for a specific period
// @Summary Calculate Profit
// @Description Calculate profit for a specific period
// @Tags profit
// @Produce json
// @Security Bearer
// @Param period query string true "Period (daily, weekly, monthly)"
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=ProfitEntry}
// @Router /api/v1/profit/calculate [get]
func (h *Handler) CalculateProfitForPeriod(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	period := c.Query("period")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if period == "" || startDateStr == "" || endDateStr == "" {
		response.BadRequest(c, "period, start_date, and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.BadRequest(c, "invalid start_date format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.BadRequest(c, "invalid end_date format")
		return
	}

	profitEntry, err := h.service.CalculateProfitForPeriod(c.Request.Context(), organizationID.(uuid.UUID), period, startDate, endDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, profitEntry, "Profit calculated successfully")
}

// GetProfitEntries retrieves profit entries for a period
// @Summary Get Profit Entries
// @Description Get profit entries for a period
// @Tags profit
// @Produce json
// @Security Bearer
// @Param period query string true "Period (daily, weekly, monthly)"
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=[]ProfitEntry}
// @Router /api/v1/profit/entries [get]
func (h *Handler) GetProfitEntries(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	period := c.Query("period")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if period == "" || startDateStr == "" || endDateStr == "" {
		response.BadRequest(c, "period, start_date, and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.BadRequest(c, "invalid start_date format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.BadRequest(c, "invalid end_date format")
		return
	}

	entries, err := h.service.GetProfitEntries(c.Request.Context(), organizationID.(uuid.UUID), period, startDate, endDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, entries, "Profit entries retrieved successfully")
}

// GetAccountBalance retrieves the balance for a specific account
// @Summary Get Account Balance
// @Description Get the balance for a specific account
// @Tags transactions
// @Produce json
// @Security Bearer
// @Param account path string true "Account name"
// @Success 200 {object} response.Response{data=gin.H}
// @Router /api/v1/transactions/accounts/{account}/balance [get]
func (h *Handler) GetAccountBalance(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	account := c.Param("account")
	if account == "" {
		response.BadRequest(c, "account parameter is required")
		return
	}

	balance, err := h.service.GetAccountBalance(c.Request.Context(), organizationID.(uuid.UUID), account)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, map[string]interface{}{
		"account": account,
		"balance": balance,
	}, "Account balance retrieved successfully")
}