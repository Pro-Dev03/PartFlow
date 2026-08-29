package users

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/middleware"
	"github.com/partflow/smart-store/pkg/response"
	"golang.org/x/crypto/bcrypt"
)

// Handler handles HTTP requests for users
type Handler struct {
	service *Service
}

// NewHandler creates a new user handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateUser handles user creation
func (h *Handler) CreateUser(c *gin.Context) {
	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Hash password
	if req.Password == "" {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Password is required", "")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req.Email, string(hashedPassword), req.FirstName, req.LastName, req.Phone, req.AvatarURL, req.IsActive)
	if err != nil {
		if err == ErrUserEmailExists {
			response.Error(c, http.StatusConflict, http.StatusConflict, "User email already exists", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to create user", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, user.ToResponse(), "User created successfully")
}

// GetUser handles user retrieval
func (h *Handler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "User not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve user", err.Error())
		return
	}

	response.Success(c, http.StatusOK, user.ToResponse(), "User retrieved successfully")
}

// ListUsers handles user listing
func (h *Handler) ListUsers(c *gin.Context) {
	var req UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	users, total, err := h.service.ListUsers(c.Request.Context(), req.Page, req.PerPage, req.Search, req.IsActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve users", err.Error())
		return
	}

	// Convert to response format
	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	response.SuccessWithPagination(c, http.StatusOK, userResponses, total, req.Page, req.PerPage, "Users retrieved successfully")
}

// UpdateUser handles user update
func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Hash password if provided
	var hashedPassword string
	if req.Password != "" {
		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to hash password", err.Error())
			return
		}
		hashedPassword = string(hashedPasswordBytes)
	}

	user, err := h.service.UpdateUser(c.Request.Context(), id, req.Email, hashedPassword, req.FirstName, req.LastName, req.Phone, req.AvatarURL, req.IsActive)
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "User not found", err.Error())
			return
		}
		if err == ErrUserEmailExists {
			response.Error(c, http.StatusConflict, http.StatusConflict, "User email already exists", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to update user", err.Error())
		return
	}

	response.Success(c, http.StatusOK, user.ToResponse(), "User updated successfully")
}

// DeleteUser handles user deletion
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	err = h.service.DeleteUser(c.Request.Context(), id)
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "User not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to delete user", err.Error())
		return
	}

	response.Success(c, http.StatusOK, nil, "User deleted successfully")
}

// ChangePassword handles password change
func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	userID := middleware.GetUserID(c)

	err := h.service.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if err == ErrCurrentPasswordIncorrect {
			response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Current password is incorrect", err.Error())
			return
		}
		if err == ErrPasswordTooShort {
			response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "New password must be at least 8 characters", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to change password", err.Error())
		return
	}

	response.Success(c, http.StatusOK, nil, "Password changed successfully")
}

// ListSubscriptionAccounts returns users with subscription metadata for admin management.
func (h *Handler) ListSubscriptionAccounts(c *gin.Context) {
	var req UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	users, total, err := h.service.ListUsers(c.Request.Context(), req.Page, req.PerPage, req.Search, req.IsActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve subscribers", err.Error())
		return
	}

	payload := make([]UserResponse, 0, len(users))
	for _, user := range users {
		payload = append(payload, user.ToResponse())
	}

	response.SuccessWithPagination(c, http.StatusOK, payload, total, req.Page, req.PerPage, "Subscribers retrieved successfully")
}

// UpdateSubscriptionStatus is an admin-only action used to activate, expire, or cancel a subscription.
func (h *Handler) UpdateSubscriptionStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req struct {
		SubscriptionStatus    string     `json:"subscription_status"`
		SubscriptionExpiresAt *time.Time `json:"subscription_expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.SubscriptionStatus == "" {
		req.SubscriptionStatus = "active"
	}

	user, err := h.service.UpdateSubscription(c.Request.Context(), id, req.SubscriptionStatus, req.SubscriptionExpiresAt)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to update subscription", err.Error())
		return
	}

	response.Success(c, http.StatusOK, user.ToResponse(), "Subscription updated successfully")
}

// RenewSubscriptionByDays renews a subscriber by x days.
func (h *Handler) RenewSubscriptionByDays(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req struct {
		Days int `json:"days" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	user, err := h.service.RenewSubscription(c.Request.Context(), id, req.Days)
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid renewal request", err.Error())
		return
	}

	response.Success(c, http.StatusOK, user.ToResponse(), "Subscription renewed successfully")
}

// GetSubscriptionSummary returns admin counts for the subscription dashboard.
func (h *Handler) GetSubscriptionSummary(c *gin.Context) {
	summary, err := h.service.GetSubscriptionSummary(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to generate subscription summary", err.Error())
		return
	}
	response.Success(c, http.StatusOK, summary, "Subscription summary retrieved successfully")
}
