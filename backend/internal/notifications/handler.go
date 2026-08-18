package notifications

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/middleware"
)

// Handler handles HTTP requests for notifications
type Handler struct {
	service *Service
}

// NewHandler creates a new notification handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateNotification handles notification creation
// @Summary Create a new notification
// @Description Create a new notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body NotificationRequest true "Notification request"
// @Success 201 {object} Notification
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/notifications [post]
func (h *Handler) CreateNotification(c *gin.Context) {
	var req NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	notification, err := h.service.CreateNotification(c.Request.Context(), organizationID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// GetNotification handles getting a notification by ID
// @Summary Get a notification
// @Description Get a notification by ID
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} Notification
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/notifications/{id} [get]
func (h *Handler) GetNotification(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	notification, err := h.service.GetNotification(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// ListNotifications handles listing notifications
// @Summary List notifications
// @Description List notifications with pagination and filters
// @Tags notifications
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param type query string false "Type filter"
// @Param status query string false "Status filter" Enums(unread, read, archived)
// @Param priority query string false "Priority filter" Enums(low, medium, high, urgent)
// @Param start_date query string false "Start date filter"
// @Param end_date query string false "End date filter"
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param sort_order query string false "Sort order" default(DESC)
// @Success 200 {object} middleware.PaginatedResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/notifications [get]
func (h *Handler) ListNotifications(c *gin.Context) {
	var req NotificationListRequest
	
	// Parse query parameters
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}
	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}
	
	req.Type = c.Query("type")
	req.Status = c.Query("status")
	req.Priority = c.Query("priority")
	req.SortBy = c.DefaultQuery("sort_by", "created_at")
	req.SortOrder = c.DefaultQuery("sort_order", "DESC")
	
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			req.StartDate = &t
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			req.EndDate = &t
		}
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	notifications, total, err := h.service.ListNotifications(c.Request.Context(), organizationID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": notifications,
		"meta": gin.H{
			"page":        req.Page,
			"per_page":    req.PerPage,
			"total":       total,
			"total_pages": (total + req.PerPage - 1) / req.PerPage,
		},
	})
}

// MarkAsRead handles marking a notification as read
// @Summary Mark notification as read
// @Description Mark a notification as read
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/notifications/{id}/read [post]
func (h *Handler) MarkAsRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	if err := h.service.MarkAsRead(c.Request.Context(), id, organizationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// MarkAllAsRead handles marking all notifications as read
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the current user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/notifications/read-all [post]
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	if err := h.service.MarkAllAsRead(c.Request.Context(), userID, organizationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteNotification handles deleting a notification
// @Summary Delete a notification
// @Description Delete a notification by ID
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/notifications/{id} [delete]
func (h *Handler) DeleteNotification(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	if err := h.service.DeleteNotification(c.Request.Context(), id, organizationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetNotificationSummary handles getting notification summary
// @Summary Get notification summary
// @Description Get notification summary for the current user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} NotificationSummary
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/notifications/summary [get]
func (h *Handler) GetNotificationSummary(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	summary, err := h.service.GetNotificationSummary(c.Request.Context(), userID, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetNotificationPreferences handles getting notification preferences
// @Summary Get notification preferences
// @Description Get notification preferences for the current user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} NotificationPreferences
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/notifications/preferences [get]
func (h *Handler) GetNotificationPreferences(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	preferences, err := h.service.GetNotificationPreferences(c.Request.Context(), userID, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preferences)
}

// UpdateNotificationPreferences handles updating notification preferences
// @Summary Update notification preferences
// @Description Update notification preferences for the current user
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body NotificationPreferences true "Notification preferences"
// @Success 200 {object} NotificationPreferences
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/notifications/preferences [put]
func (h *Handler) UpdateNotificationPreferences(c *gin.Context) {
	var preferences NotificationPreferences
	if err := c.ShouldBindJSON(&preferences); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	updatedPreferences, err := h.service.UpdateNotificationPreferences(c.Request.Context(), userID, organizationID, &preferences)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedPreferences)
}

// GetUnreadCount handles getting unread notification count
// @Summary Get unread notification count
// @Description Get count of unread notifications for the current user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} map[string]int
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/notifications/unread-count [get]
func (h *Handler) GetUnreadCount(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	count, err := h.service.GetUnreadCount(c.Request.Context(), userID, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}
