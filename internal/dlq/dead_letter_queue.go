package dlq

import (
	"context"
	"fmt"

	"github.com/longvhv/saas-shared-go/logger"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/domain"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/repository"
)

// Default maximum retry attempts before sending to DLQ
const defaultMaxRetries = 3

// DeadLetterQueue handles failed notifications
type DeadLetterQueue struct {
	repo       *repository.FailedNotificationRepository
	log        *logger.Logger
	maxRetries int
}

// NewDeadLetterQueue creates a new dead letter queue
func NewDeadLetterQueue(repo *repository.FailedNotificationRepository, log *logger.Logger) *DeadLetterQueue {
	return &DeadLetterQueue{
		repo:       repo,
		log:        log,
		maxRetries: defaultMaxRetries,
	}
}

// NewDeadLetterQueueWithRetries creates a new dead letter queue with custom max retries
func NewDeadLetterQueueWithRetries(repo *repository.FailedNotificationRepository, log *logger.Logger, maxRetries int) *DeadLetterQueue {
	return &DeadLetterQueue{
		repo:       repo,
		log:        log,
		maxRetries: maxRetries,
	}
}

// Add adds a failed notification to the DLQ
func (dlq *DeadLetterQueue) Add(ctx context.Context, notification *domain.Notification, err error) error {
	dlq.log.Warn("Adding notification to DLQ", "id", notification.ID.Hex(), "error", err)

	failed := &domain.FailedNotification{
		OriginalID: notification.ID,
		TenantID:   notification.TenantID,
		Type:       notification.Type,
		Recipient:  notification.Recipient,
		Subject:    notification.Subject,
		Body:       notification.Body,
		Payload:    notification.Payload,
		Error:      err.Error(),
		FailedAt:   notification.UpdatedAt,
		RetryCount: notification.RetryCount,
	}

	return dlq.repo.Create(ctx, failed)
}

// GetAll retrieves all failed notifications
func (dlq *DeadLetterQueue) GetAll(ctx context.Context, page, pageSize int) ([]*domain.FailedNotification, int64, error) {
	return dlq.repo.FindAll(ctx, page, pageSize)
}

// Retry retries a failed notification
func (dlq *DeadLetterQueue) Retry(ctx context.Context, id string, notificationService NotificationService) error {
	failed, err := dlq.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find notification: %w", err)
	}

	dlq.log.Info("Retrying failed notification", "id", id, "type", failed.Type)

	// Attempt to resend based on type
	switch failed.Type {
	case domain.NotificationTypeEmail:
		req := &domain.SendEmailRequest{
			TenantID: failed.TenantID,
			To:       []string{failed.Recipient},
			Subject:  failed.Subject,
			Body:     failed.Body,
		}
		err = notificationService.SendEmail(ctx, req)
	case domain.NotificationTypeSMS:
		req := &domain.SendSMSRequest{
			TenantID: failed.TenantID,
			To:       failed.Recipient,
			Message:  failed.Body,
		}
		err = notificationService.SendSMS(ctx, req)
	case domain.NotificationTypeWebhook:
		req := &domain.SendWebhookRequest{
			TenantID: failed.TenantID,
			URL:      failed.Recipient,
			Payload:  failed.Payload,
		}
		err = notificationService.SendWebhook(ctx, req)
	default:
		return fmt.Errorf("unsupported notification type: %s", failed.Type)
	}

	if err != nil {
		return fmt.Errorf("retry failed: %w", err)
	}

	// Remove from DLQ on success
	return dlq.repo.Delete(ctx, id)
}

// ShouldSendToDLQ checks if a notification should be sent to DLQ
func (dlq *DeadLetterQueue) ShouldSendToDLQ(notification *domain.Notification) bool {
	return notification.RetryCount >= dlq.maxRetries
}

// NotificationService interface for retry functionality
type NotificationService interface {
	SendEmail(ctx context.Context, req *domain.SendEmailRequest) error
	SendSMS(ctx context.Context, req *domain.SendSMSRequest) error
	SendWebhook(ctx context.Context, req *domain.SendWebhookRequest) error
}
