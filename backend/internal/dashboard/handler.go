package dashboard

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/partflow/smart-store/pkg/response"
)

type Handler struct {
	service *CachedService
}

func NewHandler(service *CachedService) *Handler {
	return &Handler{service: service}
}

// GetDashboardStats handles dashboard statistics retrieval
func (h *Handler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		// Return empty stats instead of error for empty database (offline mode)
		// This allows the app to work even with empty local database
		stats = &DashboardStats{
			TotalSales:     0,
			TotalPurchases: 0,
			TotalExpenses:  0,
			TotalRevenue:   0,
			TotalProfit:    0,
			TotalProducts:  0,
			TotalCustomers: 0,
			TotalSuppliers: 0,
			LowStockItems:  0,
			OverdueDebts:   0,
			Alerts:         []Alert{},
		}
	}

	response.OK(c, stats, "Dashboard statistics retrieved successfully")
}

func (h *Handler) GetRecentActivity(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	activityType := strings.ToLower(strings.TrimSpace(c.Query("type")))
	activity, err := h.service.GetActivity(c.Request.Context(), page, perPage, activityType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve activity", err.Error())
		return
	}

	response.OK(c, activity, "Activity retrieved successfully")
}

// GetLowStockItems handles retrieval of low stock items with details
func (h *Handler) GetLowStockItems(c *gin.Context) {
	items, err := h.service.GetLowStockItems(c.Request.Context())
	if err != nil {
		// Return empty list instead of error for empty database (offline mode)
		items = []LowStockItem{}
	}

	response.OK(c, items, "Low stock items retrieved successfully")
}

// GetOverdueDebts handles retrieval of overdue debts with details
func (h *Handler) GetOverdueDebts(c *gin.Context) {
	debts, err := h.service.GetOverdueDebts(c.Request.Context())
	if err != nil {
		// Return empty list instead of error for empty database (offline mode)
		debts = []OverdueDebtItem{}
	}

	response.OK(c, debts, "Overdue debts retrieved successfully")
}
