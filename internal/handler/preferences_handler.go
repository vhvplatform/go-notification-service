package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/middleware"
	"github.com/vhvplatform/go-notification-service/internal/repository"
	"github.com/vhvplatform/go-notification-service/internal/shared/errors"
	"github.com/vhvplatform/go-notification-service/internal/shared/logger"
)

// PreferencesHandler handles notification preferences requests
type PreferencesHandler struct {
	repo *repository.PreferencesRepository
	log  *logger.Logger
}

// NewPreferencesHandler creates a new preferences handler
func NewPreferencesHandler(repo *repository.PreferencesRepository, log *logger.Logger) *PreferencesHandler {
	return &PreferencesHandler{
		repo: repo,
		log:  log,
	}
}

// GetPreferences retrieves user notification preferences
func (h *PreferencesHandler) GetPreferences(c *gin.Context) {
	// Extract tenant_id from authenticated context
	tenantID := middleware.MustGetTenantID(c)
	userID := c.Param("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("user_id is required", nil))
		return
	}

	prefs, err := h.repo.GetByUserID(c.Request.Context(), tenantID, userID)
	if err != nil {
		h.log.Error("Failed to get preferences", "error", err, "tenant_id", tenantID, "user_id", userID)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to get preferences", err))
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// UpdatePreferences updates user notification preferences
func (h *PreferencesHandler) UpdatePreferences(c *gin.Context) {
	// Extract tenant_id from authenticated context
	tenantID := middleware.MustGetTenantID(c)
	userID := c.Param("user_id")

	var prefs domain.NotificationPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid request", err))
		return
	}

	// Set tenant_id and user_id from authenticated context and URL param
	prefs.TenantID = tenantID
	prefs.UserID = userID

	if err := h.repo.Update(c.Request.Context(), &prefs); err != nil {
		h.log.Error("Failed to update preferences", "error", err, "tenant_id", tenantID, "user_id", userID)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to update preferences", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Preferences updated successfully",
		"data":    prefs,
	})
}
