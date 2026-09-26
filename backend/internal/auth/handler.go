package auth

import (
	"errors"
	"io"
	"net"
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

const (
	refreshTokenCookieName = "partflow_refresh_token"
	refreshTokenCookieAge  = 7 * 24 * 60 * 60
)

func setRefreshTokenCookie(c *gin.Context, token string, maxAge int) {
	forwardedProto := strings.TrimSpace(strings.Split(c.GetHeader("X-Forwarded-Proto"), ",")[0])
	secure := c.Request.TLS != nil ||
		strings.EqualFold(forwardedProto, "https") ||
		strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") ||
		isLoopbackAuthHost(c.Request.Host)
	if secure {
		// The web client may run on a different origin from the cloud API.
		// SameSite=None is required for credentialed cross-site refresh calls.
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie(refreshTokenCookieName, token, maxAge, "/api/v1/auth", "", secure, true)
}

func isLoopbackAuthHost(host string) bool {
	host = strings.TrimSpace(host)
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func clearRefreshTokenCookie(c *gin.Context) {
	setRefreshTokenCookie(c, "", -1)
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
	// Subscriber accounts are created only through the protected administrator
	// route. Client-controlled configuration must never enable registration.
	c.JSON(http.StatusForbidden, gin.H{
		"error": "public registration is disabled",
		"code":  "REGISTRATION_DISABLED",
	})
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
	setRefreshTokenCookie(c, resp.RefreshToken, refreshTokenCookieAge)

	// Response format based on worktrack
	response := gin.H{
		"data": gin.H{
			"token":        resp.AccessToken,
			"access_token": resp.AccessToken,
			"expires_in":   resp.ExpiresIn,
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
	if err := c.ShouldBindJSON(&req); err != nil && err != io.EOF {
		// The browser intentionally sends an empty JSON body for cookie-based
		// refreshes. A missing refresh_token in the body is not a protocol error
		// when the HttpOnly refresh cookie is present and is the canonical source.
		if strings.Contains(err.Error(), "RefreshToken") {
			req.RefreshToken, _ = c.Cookie(refreshTokenCookieName)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		req.RefreshToken, _ = c.Cookie(refreshTokenCookieName)
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		clearRefreshTokenCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	resp, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrInactiveUser) || errors.Is(err, ErrSubscriptionExpired) || errors.Is(err, ErrSubscriptionSuspended) || errors.Is(err, ErrAccountDeleted) || errors.Is(err, ErrUserNotFound) {
			clearRefreshTokenCookie(c)
		}
		handleAuthError(c, err)
		return
	}
	setRefreshTokenCookie(c, resp.RefreshToken, refreshTokenCookieAge)

	response := gin.H{
		"data": gin.H{
			"token":        resp.AccessToken,
			"access_token": resp.AccessToken,
			"expires_in":   resp.ExpiresIn,
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
	clearRefreshTokenCookie(c)

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
	clearRefreshTokenCookie(c)

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
		var cloudErr *CloudValidationError
		if errors.As(err, &cloudErr) {
			c.JSON(cloudErr.Status, gin.H{"error": cloudErr.Error(), "code": cloudErr.Code})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "local session service unavailable", "code": "AUTH_SERVICE_UNAVAILABLE"})
		return
	}
	setRefreshTokenCookie(c, resp.RefreshToken, refreshTokenCookieAge)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"token":        resp.AccessToken,
			"access_token": resp.AccessToken,
			"expires_in":   resp.ExpiresIn,
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
		if !errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Authentication service temporarily unavailable",
				"code":  "AUTH_SERVICE_UNAVAILABLE",
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
			"code":  "ACCOUNT_DELETED",
		})
		return
	}

	// Check if user is active and subscription is valid
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "الحساب غير نشط أو أن الاشتراك منتهٍ",
			"code":  "ACCOUNT_SUSPENDED",
		})
		return
	}

	if h.service.IsSubscriptionExpired(user.SubscriptionStatus, user.SubscriptionExpiresAt) {
		code := "SUBSCRIPTION_EXPIRED"
		if strings.EqualFold(strings.TrimSpace(user.SubscriptionStatus), "suspended") {
			code = "SUBSCRIPTION_SUSPENDED"
		} else if strings.EqualFold(strings.TrimSpace(user.SubscriptionStatus), "deleted") {
			code = "ACCOUNT_DELETED"
		}
		c.JSON(http.StatusForbidden, gin.H{
			"error": "الحساب غير نشط أو أن الاشتراك منتهٍ",
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"valid": true,
			"user": gin.H{
				"id":                      user.ID.String(),
				"email":                   user.Email,
				"first_name":              user.FirstName,
				"last_name":               user.LastName,
				"phone":                   user.Phone,
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
			if parsed, err := uuid.Parse(value); err == nil {
				return parsed
			}
		}
	}
	return uuid.Nil
}

func handleAuthError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"
	code := "INTERNAL_ERROR"

	switch {
	case errors.Is(err, ErrSubscriptionExpired):
		status = http.StatusForbidden
		message = "اشتراكك منتهي، يرجى التواصل مع الإدارة لتجديد الخدمة."
		code = "SUBSCRIPTION_EXPIRED"
	case errors.Is(err, ErrSubscriptionSuspended):
		status = http.StatusForbidden
		message = "تم إيقاف الاشتراك من الإدارة."
		code = "SUBSCRIPTION_SUSPENDED"
	case errors.Is(err, ErrAccountDeleted):
		status = http.StatusUnauthorized
		message = "الحساب محذوف ولم يعد صالحًا للدخول."
		code = "ACCOUNT_DELETED"
	case errors.Is(err, ErrUserNotFound):
		status = http.StatusUnauthorized
		message = err.Error()
		code = "ACCOUNT_DELETED"
	case errors.Is(err, ErrInactiveUser):
		status = http.StatusForbidden
		message = "الحساب موقوف من الإدارة."
		code = "ACCOUNT_SUSPENDED"
	case errors.Is(err, ErrInvalidCredentials),
		errors.Is(err, ErrInvalidPassword),
		errors.Is(err, ErrUnauthorized),
		errors.Is(err, ErrInvalidToken),
		errors.Is(err, ErrTokenExpired):
		status = http.StatusUnauthorized
		message = err.Error()
		code = "INVALID_TOKEN"
	case errors.Is(err, ErrUserExists):
		status = http.StatusConflict
		message = err.Error()
	}

	c.JSON(status, gin.H{"error": message, "code": code})
}
