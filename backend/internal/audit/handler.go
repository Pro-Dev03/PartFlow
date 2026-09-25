package audit

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/response"
)

// Handler handles audit log HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new audit log handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateAuditLog creates a new audit log entry
func (h *Handler) CreateAuditLog(c *gin.Context) {
	var req AuditLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	// Override user_id from context
	req.UserID = userID.(uuid.UUID)

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	requestID := c.GetHeader("X-Request-ID")

	auditLog, err := h.service.CreateAuditLog(c.Request.Context(), &req, ipAddress, userAgent, requestID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to create audit log", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, auditLog, "Audit log created successfully")
}

// GetAuditLog retrieves an audit log by ID
func (h *Handler) GetAuditLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid audit log ID", err.Error())
		return
	}

	auditLog, err := h.service.GetAuditLog(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, http.StatusNotFound, "Audit log not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, auditLog, "Audit log retrieved successfully")
}

// ListAuditLogs retrieves audit logs with pagination and filters
func (h *Handler) ListAuditLogs(c *gin.Context) {
	var req AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	auditLogs, total, err := h.service.ListAuditLogs(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve audit logs", err.Error())
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, auditLogs, total, req.Page, req.PerPage, "Audit logs retrieved successfully")
}

// GetAuditLogSummary retrieves audit log summary statistics
func (h *Handler) GetAuditLogSummary(c *gin.Context) {
	summary, err := h.service.GetAuditLogSummary(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve audit log summary", err.Error())
		return
	}

	response.Success(c, http.StatusOK, summary, "Audit log summary retrieved successfully")
}

// GetUserAuditLogs retrieves audit logs for a specific user
func (h *Handler) GetUserAuditLogs(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	req.UserID = &userID

	auditLogs, total, err := h.service.ListAuditLogs(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve user audit logs", err.Error())
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, auditLogs, total, req.Page, req.PerPage, "User audit logs retrieved successfully")
}

// GetEntityAuditLogs retrieves audit logs for a specific entity
func (h *Handler) GetEntityAuditLogs(c *gin.Context) {
	entityID, err := uuid.Parse(c.Param("entity_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid entity ID", err.Error())
		return
	}

	entityType := c.Query("entity_type")
	if entityType == "" {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Entity type is required", "")
		return
	}

	var req AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	req.EntityID = &entityID
	req.EntityType = entityType

	auditLogs, total, err := h.service.ListAuditLogs(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve entity audit logs", err.Error())
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, auditLogs, total, req.Page, req.PerPage, "Entity audit logs retrieved successfully")
}

// ExportAuditLogs exports audit logs to CSV
func (h *Handler) ExportAuditLogs(c *gin.Context) {
	var req AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	// Set high limit for export
	req.PerPage = 10000
	req.Page = 1

	auditLogs, _, err := h.service.ListAuditLogsForExport(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve audit logs for export", err.Error())
		return
	}

	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	if err := writer.Write([]string{"ID", "Action", "EntityType", "EntityID", "Description", "Status", "UserID", "CreatedAt"}); err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to create audit export", err.Error())
		return
	}
	for _, log := range auditLogs {
		row := []string{
			safeCSVValue(log["id"]),
			safeCSVValue(log["action"]),
			safeCSVValue(log["entity_type"]),
			safeCSVValue(log["entity_id"]),
			safeCSVValue(log["description"]),
			safeCSVValue(log["status"]),
			safeCSVValue(log["user_id"]),
			safeCSVValue(log["created_at"]),
		}
		if err := writer.Write(row); err != nil {
			response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to create audit export", err.Error())
			return
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to create audit export", err.Error())
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=audit_logs.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", append([]byte{0xEF, 0xBB, 0xBF}, output.Bytes()...))
}

func safeCSVValue(value interface{}) string {
	if value == nil {
		return ""
	}
	text := fmt.Sprint(value)
	trimmed := strings.TrimLeft(text, " \t\r\n")
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + text
	}
	return text
}

// GetAuditStats retrieves audit statistics for dashboard
func (h *Handler) GetAuditStats(c *gin.Context) {
	days := c.DefaultQuery("days", "30")
	_, err := strconv.Atoi(days)
	if err != nil {
		// Use default
	}

	stats, err := h.service.GetAuditLogSummary(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve audit statistics", err.Error())
		return
	}

	response.Success(c, http.StatusOK, stats, "Audit statistics retrieved successfully")
}
