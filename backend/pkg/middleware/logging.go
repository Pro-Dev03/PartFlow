package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/logger"
)

// LoggingMiddleware provides structured logging for HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		userAgent := c.GetHeader("User-Agent")
		ipAddress := c.ClientIP()
		
		// Get request ID
		requestID := GetRequestID(c)
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set("request_id", requestID)
			c.Header("X-Request-ID", requestID)
		}

		// Get user context if available
		userID := GetUserID(c).String()
		organizationID := GetOrganizationID(c).String()

		// Create request logger
		requestLogger := logger.WithRequest(requestID, userID, organizationID, method, path)

		// Log request start
		requestLogger.Debug().
			Str("user_agent", userAgent).
			Str("ip_address", ipAddress).
			Msg("Request started")

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Log request completion
		logger.LogHTTPRequest(method, path, userAgent, ipAddress, statusCode, duration)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				requestLogger.Error().
					Str("error_type", string(err.Type)).
					Msg(err.Error())
			}
		}

		// Log response details for debugging
		if statusCode >= 400 {
			requestLogger.Warn().
				Int("status_code", statusCode).
				Int64("duration_ms", duration.Milliseconds()).
				Msg("Request completed with error")
		} else {
			requestLogger.Info().
				Int("status_code", statusCode).
				Int64("duration_ms", duration.Milliseconds()).
				Msg("Request completed successfully")
		}
	}
}

// ErrorLoggingMiddleware logs errors with detailed context
func ErrorLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for errors
		if len(c.Errors) > 0 {
			requestID := GetRequestID(c)
			userID := GetUserID(c).String()
			organizationID := GetOrganizationID(c).String()
			method := c.Request.Method
			path := c.Request.URL.Path

			errorLogger := logger.WithRequest(requestID, userID, organizationID, method, path)

			for _, err := range c.Errors {
				errorLogger.Error().
					Str("error_type", string(err.Type)).
					Interface("error_meta", err.Meta).
					Msg(err.Error())
			}
		}
	}
}

// PerformanceLoggingMiddleware logs performance metrics
func PerformanceLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		
		// Log slow requests (> 1 second)
		if duration > time.Second {
			requestID := GetRequestID(c)
			userID := GetUserID(c).String()
			organizationID := GetOrganizationID(c).String()
			method := c.Request.Method
			path := c.Request.URL.Path

			logger.Warn("Slow request detected", map[string]interface{}{
				"request_id":   requestID,
				"user_id":      userID,
				"organization_id": organizationID,
				"method":       method,
				"path":         path,
				"duration_ms":  duration.Milliseconds(),
			})
		}
	}
}

// SecurityLoggingMiddleware logs security-related events
func SecurityLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log authentication failures
		if c.Writer.Status() == 401 {
			requestID := GetRequestID(c)
			method := c.Request.Method
			path := c.Request.URL.Path
			ipAddress := c.ClientIP()
			userAgent := c.GetHeader("User-Agent")

			logger.Warn("Authentication failed", map[string]interface{}{
				"request_id":  requestID,
				"method":      method,
				"path":        path,
				"ip_address":  ipAddress,
				"user_agent":  userAgent,
			})
		}

		// Log authorization failures
		if c.Writer.Status() == 403 {
			requestID := GetRequestID(c)
			userID := GetUserID(c).String()
			organizationID := GetOrganizationID(c).String()
			method := c.Request.Method
			path := c.Request.URL.Path
			ipAddress := c.ClientIP()

			logger.Warn("Authorization failed - access denied", map[string]interface{}{
				"request_id":   requestID,
				"user_id":      userID,
				"organization_id": organizationID,
				"method":       method,
				"path":         path,
				"ip_address":   ipAddress,
			})
		}

		c.Next()
	}
}