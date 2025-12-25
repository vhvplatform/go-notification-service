package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/longvhv/saas-shared-go/errors"
	"github.com/longvhv/saas-shared-go/logger"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/domain"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/service"
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
