package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/vhvcorp/go-notification-service/internal/metrics"
	"golang.org/x/time/rate"
)

// TenantRateLimiter manages rate limiters per tenant
type TenantRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewTenantRateLimiter creates a new tenant rate limiter
func NewTenantRateLimiter(rps float64, burst int) *TenantRateLimiter {
	return &TenantRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(rps),
		burst:    burst,
	}
}

// GetLimiter returns the rate limiter for a specific tenant
func (rl *TenantRateLimiter) GetLimiter(tenantID string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[tenantID]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check after acquiring write lock
		limiter, exists = rl.limiters[tenantID]
		if !exists {
			limiter = rate.NewLimiter(rl.rate, rl.burst)
			rl.limiters[tenantID] = limiter
		}
		rl.mu.Unlock()
	}

	return limiter
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rl *TenantRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to extract tenant_id from query parameter first (doesn't consume body)
		tenantID := c.Query("tenant_id")
		
		// If not in query, try from form data
		if tenantID == "" {
			tenantID = c.PostForm("tenant_id")
		}
		
		// If still empty, try from JSON body (use peek method to not consume)
		if tenantID == "" {
			var req struct {
				TenantID string `json:"tenant_id"`
			}
			// ShouldBindBodyWith allows binding without consuming the body
			if err := c.ShouldBindBodyWith(&req, binding.JSON); err == nil {
				tenantID = req.TenantID
			}
		}
		
		// If still empty, allow through (will fail validation later)
		if tenantID == "" {
			c.Next()
			return
		}

		limiter := rl.GetLimiter(tenantID)

		if !limiter.Allow() {
			metrics.RateLimitExceeded.WithLabelValues(tenantID).Inc()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
