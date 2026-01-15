package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/middleware"
	"github.com/vhvplatform/go-notification-service/internal/service"
	"github.com/vhvplatform/go-notification-service/internal/shared/errors"
	"github.com/vhvplatform/go-notification-service/internal/shared/logger"
)

// SMSHandler handles SMS notification requests
type SMSHandler struct {
	service *service.NotificationService
	log     *logger.Logger
}

// NewSMSHandler creates a new SMS handler
func NewSMSHandler(service *service.NotificationService, log *logger.Logger) *SMSHandler {
	return &SMSHandler{
		service: service,
		log:     log,
	}
}

// SendSMS handles SMS notification requests
func (h *SMSHandler) SendSMS(c *gin.Context) {
	// Extract tenant_id from authenticated context
	tenantID := middleware.MustGetTenantID(c)

	var req domain.SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid request", err))
		return
	}

	// Set tenant_id from authenticated context
	req.TenantID = tenantID

	if err := h.service.SendSMS(c.Request.Context(), &req); err != nil {
		h.log.Error("Failed to send SMS", "error", err, "tenant_id", tenantID)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to send SMS", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SMS sent successfully",
	})
}
