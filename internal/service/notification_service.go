package service

import (
	"context"

	"github.com/longvhv/saas-framework-go/pkg/logger"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/domain"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/repository"
)

// NotificationService handles notification business logic
type NotificationService struct {
	notifRepo      *repository.NotificationRepository
	emailService   *EmailService
	webhookService *WebhookService
	log            *logger.Logger
}

// NewNotificationService creates a new notification service
func NewNotificationService(notifRepo *repository.NotificationRepository, emailService *EmailService, webhookService *WebhookService, log *logger.Logger) *NotificationService {
	return &NotificationService{
		notifRepo:      notifRepo,
		emailService:   emailService,
		webhookService: webhookService,
		log:            log,
	}
}

// SendEmail sends an email notification
func (s *NotificationService) SendEmail(ctx context.Context, req *domain.SendEmailRequest) error {
	s.log.Info("Sending email notification", "tenant_id", req.TenantID, "recipients", len(req.To))
	return s.emailService.SendEmail(ctx, req)
}

// SendWebhook sends a webhook notification
func (s *NotificationService) SendWebhook(ctx context.Context, req *domain.SendWebhookRequest) error {
	s.log.Info("Sending webhook notification", "tenant_id", req.TenantID, "url", req.URL)
	return s.webhookService.SendWebhook(ctx, req)
}

// GetNotifications retrieves notifications with pagination
func (s *NotificationService) GetNotifications(ctx context.Context, req *domain.GetNotificationsRequest) ([]*domain.Notification, int64, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.notifRepo.FindByTenantID(ctx, req.TenantID, req.Type, req.Status, page, pageSize)
}

// GetNotification retrieves a single notification by ID
func (s *NotificationService) GetNotification(ctx context.Context, id string) (*domain.Notification, error) {
	return s.notifRepo.FindByID(ctx, id)
}

// ProcessEvent processes events from RabbitMQ
func (s *NotificationService) ProcessEvent(ctx context.Context, event *domain.Event) error {
	s.log.Info("Processing event", "type", event.Type, "tenant_id", event.TenantID)

	switch event.Type {
	case domain.EventUserRegistered:
		return s.handleUserRegistered(ctx, event)
	case domain.EventUserPasswordReset:
		return s.handlePasswordReset(ctx, event)
	case domain.EventTenantCreated:
		return s.handleTenantCreated(ctx, event)
	case domain.EventPaymentCompleted:
		return s.handlePaymentCompleted(ctx, event)
	default:
		s.log.Warn("Unknown event type", "type", event.Type)
		return nil
	}
}

// handleUserRegistered handles user registration events
func (s *NotificationService) handleUserRegistered(ctx context.Context, event *domain.Event) error {
	email, ok := event.Email
	if !ok || email == "" {
		s.log.Warn("User registration event missing email", "event", event)
		return nil
	}

	req := &domain.SendEmailRequest{
		TenantID: event.TenantID,
		To:       []string{email},
		Subject:  "Welcome to our platform!",
		Body:     "Thank you for registering. We're excited to have you on board!",
		IsHTML:   false,
	}

	return s.emailService.SendEmail(ctx, req)
}

// handlePasswordReset handles password reset events
func (s *NotificationService) handlePasswordReset(ctx context.Context, event *domain.Event) error {
	email, ok := event.Email
	if !ok || email == "" {
		return nil
	}

	resetToken := ""
	if event.Data != nil {
		if token, ok := event.Data["reset_token"].(string); ok {
			resetToken = token
		}
	}

	req := &domain.SendEmailRequest{
		TenantID: event.TenantID,
		To:       []string{email},
		Subject:  "Password Reset Request",
		Body:     "Your password reset token: " + resetToken,
		IsHTML:   false,
	}

	return s.emailService.SendEmail(ctx, req)
}

// handleTenantCreated handles tenant creation events
func (s *NotificationService) handleTenantCreated(ctx context.Context, event *domain.Event) error {
	// Send webhook or email notification for tenant creation
	s.log.Info("Tenant created", "tenant_id", event.TenantID)
	return nil
}

// handlePaymentCompleted handles payment completion events
func (s *NotificationService) handlePaymentCompleted(ctx context.Context, event *domain.Event) error {
	email, ok := event.Email
	if !ok || email == "" {
		return nil
	}

	req := &domain.SendEmailRequest{
		TenantID: event.TenantID,
		To:       []string{email},
		Subject:  "Payment Confirmation",
		Body:     "Your payment has been processed successfully!",
		IsHTML:   false,
	}

	return s.emailService.SendEmail(ctx, req)
}
