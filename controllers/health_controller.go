package controllers

import (
	"net/http"
	"time"

	"suberes_golang/config"

	"github.com/gin-gonic/gin"
)

// HealthController provides standard health probe endpoints used by:
//   - Docker HEALTHCHECK
//   - Load balancers (readiness / liveness probes)
//   - Uptime monitoring services
type HealthController struct{}

// Health returns a simple liveness response.
// GET /health  — no authentication required.
func (h *HealthController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "suberes-api",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// Liveness reports whether the process is alive (Kubernetes-style).
// GET /live  — no authentication required.
func (h *HealthController) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

// Readiness checks all critical dependencies before reporting ready.
// Returns 503 when the database is unreachable so load balancers can stop
// sending traffic during cold-start or DB failover.
// GET /ready  — no authentication required.
func (h *HealthController) Readiness(c *gin.Context) {
	sqlDB, err := config.DB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"checks": gin.H{"database": "unreachable"},
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"checks": gin.H{"database": "ping_failed"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": gin.H{"database": "ok"},
	})
}
