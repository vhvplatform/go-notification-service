package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/service"
	"github.com/vhvplatform/go-notification-service/internal/shared/errors"
	"github.com/vhvplatform/go-notification-service/internal/shared/logger"
)

// NotificationHandler handles HTTP requests for notifications
type NotificationHandler struct {
	service *service.NotificationService
	log     *logger.Logger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(service *service.NotificationService, log *logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		log:     log,
	}
}

// SendEmail handles email notification requests
func (h *NotificationHandler) SendEmail(c *gin.Context) {
	var req domain.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid request", err))
		return
	}

	if err := h.service.SendEmail(c.Request.Context(), &req); err != nil {
		h.log.Error("Failed to send email", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to send email", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email sent successfully",
	})
}

// SendWebhook handles webhook notification requests
func (h *NotificationHandler) SendWebhook(c *gin.Context) {
	var req domain.SendWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid request", err))
		return
	}

	if err := h.service.SendWebhook(c.Request.Context(), &req); err != nil {
		h.log.Error("Failed to send webhook", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to send webhook", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook sent successfully",
	})
}

// GetNotifications retrieves notification history
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	var req domain.GetNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid request", err))
		return
	}

	notifications, total, err := h.service.GetNotifications(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("Failed to get notifications", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to get notifications", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      notifications,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetNotification retrieves a single notification by ID
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("ID is required", nil))
		return
	}

	notification, err := h.service.GetNotification(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get notification", "error", err, "id", id)
		c.JSON(http.StatusNotFound, errors.NewNotFoundError("Notification not found", err))
		return
	}

	c.JSON(http.StatusOK, notification)
}
