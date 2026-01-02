package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vhvplatform/go-notification-service/internal/dlq"
	"github.com/vhvplatform/go-notification-service/internal/service"
	"github.com/vhvplatform/go-notification-service/internal/shared/errors"
	"github.com/vhvplatform/go-notification-service/internal/shared/logger"
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

// GetFailedNotifications godoc
// @Summary Get failed notifications
// @Description Get list of failed notifications from DLQ
// @Tags dlq
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} map[string]interface{} "List of failed notifications"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/dlq [get]
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

// RetryNotification godoc
// @Summary Retry failed notification
// @Description Retry sending a failed notification
// @Tags dlq
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]interface{} "Notification retried"
// @Failure 404 {object} map[string]interface{} "Notification not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/dlq/{id}/retry [post]
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
