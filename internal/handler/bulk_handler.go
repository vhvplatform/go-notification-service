package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vhvcorp/go-notification-service/internal/domain"
	"github.com/vhvcorp/go-notification-service/internal/service"
	"github.com/vhvcorp/go-shared/errors"
	"github.com/vhvcorp/go-shared/logger"
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

// SendBulkEmail handles bulk email requests
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
