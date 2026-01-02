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

// SendEmail godoc
// @Summary Send email notification
// @Description Send an email notification to recipients
// @Tags notifications
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param email body object true "Email request"
// @Success 200 {object} map[string]interface{} "Email sent successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/notifications/email [post]
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

// SendWebhook godoc
// @Summary Send webhook notification
// @Description Send a webhook notification to a specified URL
// @Tags notifications
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param webhook body object true "Webhook request"
// @Success 200 {object} map[string]interface{} "Webhook sent successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/notifications/webhook [post]
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

// GetNotifications godoc
// @Summary Get notifications
// @Description Get list of notifications with pagination
// @Tags notifications
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} map[string]interface{} "List of notifications"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/notifications [get]
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

// GetNotification godoc
// @Summary Get notification by ID
// @Description Get details of a specific notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]interface{} "Notification details"
// @Failure 404 {object} map[string]interface{} "Notification not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/notifications/{id} [get]
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
