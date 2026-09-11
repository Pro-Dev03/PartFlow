package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Handler struct {
	service *Service
	db      *sqlx.DB
}

func NewHandler(service *Service, db *sqlx.DB) *Handler {
	return &Handler{service: service, db: db}
}

// RegisterRoutes registers auth routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
		auth.POST("/change-password", h.ChangePassword)
		auth.POST("/cloud-session", h.CloudSession)
		auth.POST("/validate", h.ValidateSubscription)
		// auth.POST("/password-reset", h.RequestPasswordReset)
		// auth.POST("/password-reset/confirm", h.ResetPassword)
	}

	users := router.Group("/users")
	{
		users.GET("/me", h.GetCurrentUser)
	}
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	// Subscriber accounts are created and managed from the administrator
	// tooling. Keep public registration disabled by default; it can be enabled
	// explicitly for a development or invite-based deployment.
	allowPublicRegistration := strings.EqualFold(strings.TrimSpace(os.Getenv("PARTFLOW_ALLOW_PUBLIC_REGISTRATION")), "true") ||
		os.Getenv("PARTFLOW_ALLOW_PUBLIC_REGISTRATION") == "1"
	if !allowPublicRegistration {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "public registration is disabled",
			"code":  "REGISTRATION_DISABLED",
		})
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login handles admin login (based on worktrack)
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Response format based on worktrack
	response := gin.H{
		"data": gin.H{
			"token":         resp.AccessToken,
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"expires_in":    resp.ExpiresIn,
			"user": gin.H{
				"id":                      resp.User.ID.String(),
				"email":                   resp.User.Email,
				"first_name":              resp.User.FirstName,
				"last_name":               resp.User.LastName,
				"phone":                   resp.User.Phone,
				"is_active":               resp.User.IsActive,
				"subscription_status":     resp.User.SubscriptionStatus,
				"subscription_expires_at": resp.User.SubscriptionExpiresAt,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	response := gin.H{
		"data": gin.H{
			"token":         resp.AccessToken,
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"expires_in":    resp.ExpiresIn,
			"user": gin.H{
				"id":                      resp.User.ID.String(),
				"email":                   resp.User.Email,
				"first_name":              resp.User.FirstName,
				"last_name":               resp.User.LastName,
				"phone":                   resp.User.Phone,
				"is_active":               resp.User.IsActive,
				"subscription_status":     resp.User.SubscriptionStatus,
				"subscription_expires_at": resp.User.SubscriptionExpiresAt,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	userID := getUserIDFromContext(c)

	if err := h.service.Logout(c.Request.Context(), userID); err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// ChangePassword handles password change
func (h *Handler) ChangePassword(c *gin.Context) {
	userID := getUserIDFromContext(c)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// GetCurrentUser returns the current authenticated user
func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID := getUserIDFromContext(c)

	user, err := h.service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// CloudSession creates a local JWT session from a cloud access token
func (h *Handler) CloudSession(c *gin.Context) {
	var req CloudSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateCloudSession(c.Request.Context(), req.CloudToken)
	if err != nil {
		status := http.StatusUnauthorized
		if strings.Contains(err.Error(), "cloud auth is not configured") {
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"token":         resp.AccessToken,
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"expires_in":    resp.ExpiresIn,
			"user": gin.H{
				"id":                      resp.User.ID.String(),
				"email":                   resp.User.Email,
				"first_name":              resp.User.FirstName,
				"last_name":               resp.User.LastName,
				"phone":                   resp.User.Phone,
				"is_active":               resp.User.IsActive,
				"subscription_status":     resp.User.SubscriptionStatus,
				"subscription_expires_at": resp.User.SubscriptionExpiresAt,
			},
		},
	})
}

// ValidateSubscription confirms the account is currently permitted by the
// cloud. The authentication middleware has already checked account activity
// and subscription expiry.
func (h *Handler) ValidateSubscription(c *gin.Context) {
	userID := getUserIDFromContext(c)
	
	// If no user ID from context, try to get from JWT token directly
	if userID == uuid.Nil {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := h.service.ValidateToken(c.Request.Context(), tokenString)
			if err == nil && claims != nil {
				userID, _ = uuid.Parse(claims.UserID)
			}
		}
	}
	
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or missing token",
			"code":  "INVALID_TOKEN",
		})
		return
	}
	
	user, err := h.service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}
	
	// Check if user is active and subscription is valid
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "الحساب غير نشط أو أن الاشتراك منتهٍ",
			"code":  "SUBSCRIPTION_EXPIRED",
		})
		return
	}
	
	if h.service.IsSubscriptionExpired(user.SubscriptionStatus, user.SubscriptionExpiresAt) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "الحساب غير نشط أو أن الاشتراك منتهٍ",
			"code":  "SUBSCRIPTION_EXPIRED",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"valid":                   true,
			"user": gin.H{
				"id":                      user.ID.String(),
				"email":                   user.Email,
				"first_name":              user.FirstName,
				"last_name":               user.LastName,
				"is_active":               user.IsActive,
				"subscription_status":     user.SubscriptionStatus,
				"subscription_expires_at": user.SubscriptionExpiresAt,
			},
			"subscription_status":     user.SubscriptionStatus,
			"subscription_expires_at": user.SubscriptionExpiresAt,
		},
	})
}

// Helper functions

func getUserIDFromContext(c *gin.Context) uuid.UUID {
	if userID, exists := c.Get("user_id"); exists {
		if parsed, ok := userID.(uuid.UUID); ok {
			return parsed
		}
		if value, ok := userID.(string); ok {
			return uuid.MustParse(value)
		}
	}
	return uuid.Nil
}

func handleAuthError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"

	switch {
	case errors.Is(err, ErrUserNotFound):
		status = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, ErrInvalidCredentials),
		errors.Is(err, ErrInvalidPassword),
		errors.Is(err, ErrInactiveUser),
		errors.Is(err, ErrUnauthorized),
		errors.Is(err, ErrInvalidToken),
		errors.Is(err, ErrTokenExpired):
		status = http.StatusUnauthorized
		message = err.Error()
	case errors.Is(err, ErrUserExists):
		status = http.StatusConflict
		message = err.Error()
	}

	c.JSON(status, gin.H{"error": message})
}
