package webhook

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vhvcorp/go-notification-service/internal/domain"
	"github.com/vhvcorp/go-notification-service/internal/metrics"
	"github.com/vhvcorp/go-notification-service/internal/repository"
	"github.com/vhvcorp/go-notification-service/internal/shared/logger"
)

// BounceHandler handles email bounce webhooks
type BounceHandler struct {
	repo *repository.BounceRepository
	log  *logger.Logger
}

// BounceEvent represents a bounce event from an email provider
type BounceEvent struct {
	Type       string    `json:"type"` // bounce, complaint
	Email      string    `json:"email"`
	Timestamp  time.Time `json:"timestamp"`
	Reason     string    `json:"reason"`
	BounceType string    `json:"bounce_type"` // hard, soft
}

// NewBounceHandler creates a new bounce handler
func NewBounceHandler(repo *repository.BounceRepository, log *logger.Logger) *BounceHandler {
	return &BounceHandler{
		repo: repo,
		log:  log,
	}
}

// HandleSESWebhook handles AWS SES bounce webhooks
func (h *BounceHandler) HandleSESWebhook(c *gin.Context) {
	var event BounceEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.log.Error("Invalid bounce event", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	h.log.Info("Received bounce event", "email", event.Email, "type", event.Type)

	// Record bounce
	bounce := &domain.EmailBounce{
		Email:     event.Email,
		Type:      event.BounceType,
		Reason:    event.Reason,
		Timestamp: event.Timestamp,
	}

	if err := h.repo.Create(c.Request.Context(), bounce); err != nil {
		h.log.Error("Failed to record bounce", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process bounce"})
		return
	}

	// Update metrics
	metrics.EmailBounces.WithLabelValues(event.BounceType).Inc()

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// HandleSendGridWebhook handles SendGrid bounce webhooks
func (h *BounceHandler) HandleSendGridWebhook(c *gin.Context) {
	var events []BounceEvent
	if err := c.ShouldBindJSON(&events); err != nil {
		h.log.Error("Invalid SendGrid event", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	for _, event := range events {
		h.log.Info("Received SendGrid bounce event", "email", event.Email, "type", event.Type)

		bounce := &domain.EmailBounce{
			Email:     event.Email,
			Type:      event.BounceType,
			Reason:    event.Reason,
			Timestamp: event.Timestamp,
		}

		if err := h.repo.Create(c.Request.Context(), bounce); err != nil {
			h.log.Error("Failed to record bounce", "error", err)
			continue
		}

		metrics.EmailBounces.WithLabelValues(event.BounceType).Inc()
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
