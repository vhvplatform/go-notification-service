package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/longvhv/saas-framework-go/pkg/errors"
	"github.com/longvhv/saas-framework-go/pkg/logger"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/dlq"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/service"
)

// DLQHandler handles dead letter queue operations
type DLQHandler struct {
	dlq     *dlq.DeadLetterQueue
	service *service.NotificationService
	log     *logger.Logger
}

// NewDLQHandler creates a new DLQ handler
func NewDLQHandler(dlq *dlq.DeadLetterQueue, service *service.NotificationService, log *logger.Logger) *DLQHandler {
	return &DLQHandler{
		dlq:     dlq,
		service: service,
		log:     log,
	}
}

// GetFailedNotifications retrieves failed notifications from DLQ
func (h *DLQHandler) GetFailedNotifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	failed, total, err := h.dlq.GetAll(c.Request.Context(), page, pageSize)
	if err != nil {
		h.log.Error("Failed to get failed notifications", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to get failed notifications", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      failed,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RetryNotification retries a failed notification
func (h *DLQHandler) RetryNotification(c *gin.Context) {
	id := c.Param("id")

	if err := h.dlq.Retry(c.Request.Context(), id, h.service); err != nil {
		h.log.Error("Failed to retry notification", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to retry notification", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification retried successfully",
	})
}
