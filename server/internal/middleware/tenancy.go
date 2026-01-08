package middleware

import (
	"context"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// TenantIDKey is the context key for storing tenant ID
	TenantIDKey ContextKey = "tenant_id"

	// TenantIDHeader is the HTTP header name for tenant ID
	TenantIDHeader = "X-Tenant-ID"

	// AuditTenantIDKey is the key for audit logging
	AuditTenantIDKey = "audit_tenant_id"

	// tenantIDPattern defines allowed characters for tenant IDs
	tenantIDPattern = `^[a-zA-Z0-9_-]+$`
)

var (
	// tenantIDRegex is the compiled regex for tenant ID validation
	tenantIDRegex = regexp.MustCompile(tenantIDPattern)
)

// TenancyMiddleware extracts the X-Tenant-ID header and validates tenant isolation
// This middleware must be applied to all tenant-aware routes
// Returns 400 Bad Request if X-Tenant-ID header is missing
func TenancyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract X-Tenant-ID header
		tenantID := c.GetHeader(TenantIDHeader)

		// Validate tenant ID is present
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Missing tenant identifier",
				"message": "X-Tenant-ID header is required for all tenant operations",
				"code":    "TENANT_ID_REQUIRED",
			})
			c.Abort()
			return
		}

		// Validate tenant ID format (basic validation)
		if len(tenantID) < 3 || len(tenantID) > 128 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid tenant identifier",
				"message": "X-Tenant-ID must be between 3 and 128 characters",
				"code":    "INVALID_TENANT_ID",
			})
			c.Abort()
			return
		}

		// Validate tenant ID contains only allowed characters
		if !tenantIDRegex.MatchString(tenantID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid tenant identifier format",
				"message": "X-Tenant-ID must contain only alphanumeric characters, hyphens, and underscores",
				"code":    "INVALID_TENANT_ID_FORMAT",
			})
			c.Abort()
			return
		}

		// Store tenant ID in Gin context
		c.Set(string(TenantIDKey), tenantID)

		// Also store in request context for use in non-Gin code
		ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenantID)
		c.Request = c.Request.WithContext(ctx)

		// Log tenant ID for audit trail
		c.Set(AuditTenantIDKey, tenantID)

		// Continue to next handler
		c.Next()
	}
}

// GetTenantID retrieves the tenant ID from Gin context
// Returns empty string if tenant ID is not found
func GetTenantID(c *gin.Context) string {
	if tenantID, exists := c.Get(string(TenantIDKey)); exists {
		if tid, ok := tenantID.(string); ok {
			return tid
		}
	}
	return ""
}

// GetTenantIDFromContext retrieves tenant ID from standard context
// Useful for non-Gin code (database queries, background jobs, etc.)
func GetTenantIDFromContext(ctx context.Context) string {
	if tenantID := ctx.Value(TenantIDKey); tenantID != nil {
		if tid, ok := tenantID.(string); ok {
			return tid
		}
	}
	return ""
}

// MustGetTenantID retrieves tenant ID from context and panics if not found
// IMPORTANT: Only use this in handlers where TenancyMiddleware is guaranteed to be applied.
// This is a programming error if tenant ID is missing at this point, not a runtime error.
// For optional tenant ID retrieval, use GetTenantID() instead.
func MustGetTenantID(c *gin.Context) string {
	tenantID := GetTenantID(c)
	if tenantID == "" {
		// This panic indicates a middleware configuration error, not a user error
		panic("tenant ID not found in context - ensure TenancyMiddleware is applied to this route")
	}
	return tenantID
}

// GetTenantIDOrError retrieves tenant ID from context or returns an HTTP error
// Use this when you want graceful error handling instead of panicking
func GetTenantIDOrError(c *gin.Context) (string, bool) {
	tenantID := GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing tenant context",
			"message": "Tenant ID not found in request context",
			"code":    "TENANT_CONTEXT_MISSING",
		})
		return "", false
	}
	return tenantID, true
}
