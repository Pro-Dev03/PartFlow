package users

import (
	"net/http"

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

	organizationID := middleware.GetOrganizationID(c)

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

	user, err := h.service.CreateUser(c.Request.Context(), organizationID, req.Email, string(hashedPassword), req.FirstName, req.LastName, req.Phone, req.AvatarURL, req.RoleID, req.IsActive)
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

	organizationID := middleware.GetOrganizationID(c)

	user, err := h.service.GetUser(c.Request.Context(), id, organizationID)
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

	organizationID := middleware.GetOrganizationID(c)

	users, total, err := h.service.ListUsers(c.Request.Context(), organizationID, req.Page, req.PerPage, req.Search, req.IsActive, req.RoleID)
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

	organizationID := middleware.GetOrganizationID(c)

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

	user, err := h.service.UpdateUser(c.Request.Context(), id, organizationID, req.Email, hashedPassword, req.FirstName, req.LastName, req.Phone, req.AvatarURL, req.RoleID, req.IsActive)
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

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.DeleteUser(c.Request.Context(), id, organizationID)
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

// AssignRole handles role assignment to user
func (h *Handler) AssignRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.AssignRole(c.Request.Context(), id, organizationID, req.RoleID)
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "User not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to assign role", err.Error())
		return
	}

	response.Success(c, http.StatusOK, nil, "Role assigned successfully")
}

// RemoveRole handles role removal from user
func (h *Handler) RemoveRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.RemoveRole(c.Request.Context(), id, organizationID, req.RoleID)
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "User not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to remove role", err.Error())
		return
	}

	response.Success(c, http.StatusOK, nil, "Role removed successfully")
}

// GetUserRoles handles getting user roles
func (h *Handler) GetUserRoles(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	roles, err := h.service.GetUserRoles(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "User not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to get user roles", err.Error())
		return
	}

	response.Success(c, http.StatusOK, roles, "User roles retrieved successfully")
}
