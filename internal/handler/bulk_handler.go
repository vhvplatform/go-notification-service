package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/service"
	"github.com/vhvplatform/go-notification-service/internal/shared/errors"
	"github.com/vhvplatform/go-notification-service/internal/shared/logger"
)

// BulkHandler handles bulk notification operations
type BulkHandler struct {
	bulkEmailService *service.BulkEmailService
	log              *logger.Logger
}

// NewBulkHandler creates a new bulk handler
func NewBulkHandler(bulkEmailService *service.BulkEmailService, log *logger.Logger) *BulkHandler {
	return &BulkHandler{
		bulkEmailService: bulkEmailService,
		log:              log,
	}
}

// SendBulkEmail godoc
// @Summary Send bulk email notifications
// @Description Send email notifications to multiple recipients
// @Tags notifications
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param bulk body object true "Bulk email request"
// @Success 202 {object} map[string]interface{} "Bulk email accepted"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/notifications/bulk/email [post]
func (h *BulkHandler) SendBulkEmail(c *gin.Context) {
	var req domain.BulkEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid request", err))
		return
	}

	if err := h.bulkEmailService.SendBulk(c.Request.Context(), &req); err != nil {
		h.log.Error("Failed to queue bulk emails", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to queue bulk emails", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Bulk emails queued successfully",
		"count":      len(req.Recipients),
		"queue_size": h.bulkEmailService.QueueSize(),
	})
}
