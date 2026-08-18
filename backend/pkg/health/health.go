package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type HealthChecker struct {
	db *sqlx.DB
}

func NewHealthChecker(db *sqlx.DB) *HealthChecker {
	return &HealthChecker{db: db}
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
	Version   string            `json:"version"`
}

// Check performs a health check
func (h *HealthChecker) Check(c *gin.Context) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Services:  make(map[string]string),
		Version:   "1.0.0",
	}

	// Check database
	if err := h.db.Ping(); err != nil {
		response.Status = "degraded"
		response.Services["database"] = "unhealthy"
	} else {
		response.Services["database"] = "healthy"
	}

	// Check other services can be added here
	// Redis, external APIs, etc.

	if response.Status == "ok" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// Readiness checks if the application is ready to serve traffic
func (h *HealthChecker) Readiness(c *gin.Context) {
	response := HealthResponse{
		Status:    "ready",
		Timestamp: time.Now(),
		Services:  make(map[string]string),
		Version:   "1.0.0",
	}

	// Check database readiness
	if err := h.db.Ping(); err != nil {
		response.Status = "not_ready"
		response.Services["database"] = "not_ready"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response.Services["database"] = "ready"
	c.JSON(http.StatusOK, response)
}

// Liveness checks if the application is alive
func (h *HealthChecker) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}
