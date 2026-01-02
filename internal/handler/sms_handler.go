package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vhvplatform/go-notification-service/internal/domain"
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

// SendSMS godoc
// @Summary Send SMS notification
// @Description Send an SMS notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param sms body object true "SMS request"
// @Success 200 {object} map[string]interface{} "SMS sent successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/notifications/sms [post]
func (h *SMSHandler) SendSMS(c *gin.Context) {
	var req domain.SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid request", err))
		return
	}

	if err := h.service.SendSMS(c.Request.Context(), &req); err != nil {
		h.log.Error("Failed to send SMS", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to send SMS", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SMS sent successfully",
	})
}
