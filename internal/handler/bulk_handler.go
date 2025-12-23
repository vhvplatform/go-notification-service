package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/longvhv/saas-framework-go/pkg/errors"
	"github.com/longvhv/saas-framework-go/pkg/logger"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/domain"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/service"
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
