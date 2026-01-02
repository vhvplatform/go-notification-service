package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vhvplatform/go-notification-service/internal/domain"
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

// GetPreferences godoc
// @Summary Get user notification preferences
// @Description Get notification preferences for a user
// @Tags preferences
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} map[string]interface{} "User preferences"
// @Failure 404 {object} map[string]interface{} "Preferences not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/preferences/{user_id} [get]
func (h *PreferencesHandler) GetPreferences(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	userID := c.Param("user_id")

	if tenantID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("tenant_id and user_id are required", nil))
		return
	}

	prefs, err := h.repo.GetByUserID(c.Request.Context(), tenantID, userID)
	if err != nil {
		h.log.Error("Failed to get preferences", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to get preferences", err))
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// UpdatePreferences godoc
// @Summary Update user notification preferences
// @Description Update notification preferences for a user
// @Tags preferences
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param user_id path string true "User ID"
// @Param preferences body object true "Preferences update"
// @Success 200 {object} map[string]interface{} "Preferences updated"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/preferences/{user_id} [put]
func (h *PreferencesHandler) UpdatePreferences(c *gin.Context) {
	userID := c.Param("user_id")

	var prefs domain.NotificationPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid request", err))
		return
	}

	prefs.UserID = userID

	if err := h.repo.Update(c.Request.Context(), &prefs); err != nil {
		h.log.Error("Failed to update preferences", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to update preferences", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Preferences updated successfully",
		"data":    prefs,
	})
}
