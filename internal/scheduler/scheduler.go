package scheduler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/longvhv/saas-shared-go/logger"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/domain"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/repository"
	"github.com/robfig/cron/v3"
)

// NotificationScheduler manages scheduled notifications
type NotificationScheduler struct {
	cron    *cron.Cron
	service SchedulerService
	repo    *repository.ScheduledNotificationRepository
	log     *logger.Logger
	entries map[string]cron.EntryID // Maps notification ID to cron entry ID
}

// SchedulerService interface for notification operations
type SchedulerService interface {
	SendEmail(ctx context.Context, req *domain.SendEmailRequest) error
	SendSMS(ctx context.Context, req *domain.SendSMSRequest) error
	SendWebhook(ctx context.Context, req *domain.SendWebhookRequest) error
}

// NewNotificationScheduler creates a new notification scheduler
func NewNotificationScheduler(service SchedulerService, repo *repository.ScheduledNotificationRepository, log *logger.Logger) *NotificationScheduler {
	return &NotificationScheduler{
		cron:    cron.New(),
		service: service,
		repo:    repo,
		log:     log,
		entries: make(map[string]cron.EntryID),
	}
}

// Start starts the scheduler and loads active schedules
func (s *NotificationScheduler) Start() error {
	s.log.Info("Starting notification scheduler")

	// Load all active scheduled notifications
	ctx := context.Background()
	scheduled, err := s.repo.FindActive(ctx)
	if err != nil {
		return err
	}

	// Register each scheduled notification
	for _, sched := range scheduled {
		if err := s.registerSchedule(sched); err != nil {
			s.log.Error("Failed to register schedule", "error", err, "id", sched.ID)
		}
	}

	s.cron.Start()
	s.log.Info("Notification scheduler started", "active_schedules", len(scheduled))
	return nil
}

// Stop stops the scheduler
func (s *NotificationScheduler) Stop() {
	s.log.Info("Stopping notification scheduler")
	s.cron.Stop()
}

// registerSchedule registers a scheduled notification with cron
func (s *NotificationScheduler) registerSchedule(sched *domain.ScheduledNotification) error {
	entryID, err := s.cron.AddFunc(sched.Schedule, func() {
		s.executeSchedule(sched)
	})

	if err != nil {
		return err
	}

	s.entries[sched.ID.Hex()] = entryID
	s.log.Info("Registered schedule", "id", sched.ID.Hex(), "schedule", sched.Schedule, "type", sched.Type)
	return nil
}

// executeSchedule executes a scheduled notification
func (s *NotificationScheduler) executeSchedule(sched *domain.ScheduledNotification) {
	ctx := context.Background()
	s.log.Info("Executing scheduled notification", "id", sched.ID.Hex(), "type", sched.Type)

	var err error
	switch sched.Type {
	case domain.NotificationTypeEmail:
		req, parseErr := s.parseEmailRequest(sched.Request)
		if parseErr != nil {
			s.log.Error("Failed to parse email request", "error", parseErr, "id", sched.ID.Hex())
			return
		}
		err = s.service.SendEmail(ctx, req)

	case domain.NotificationTypeSMS:
		req, parseErr := s.parseSMSRequest(sched.Request)
		if parseErr != nil {
			s.log.Error("Failed to parse SMS request", "error", parseErr, "id", sched.ID.Hex())
			return
		}
		err = s.service.SendSMS(ctx, req)

	case domain.NotificationTypeWebhook:
		req, parseErr := s.parseWebhookRequest(sched.Request)
		if parseErr != nil {
			s.log.Error("Failed to parse webhook request", "error", parseErr, "id", sched.ID.Hex())
			return
		}
		err = s.service.SendWebhook(ctx, req)

	default:
		s.log.Warn("Unknown notification type", "type", sched.Type, "id", sched.ID.Hex())
		return
	}

	if err != nil {
		s.log.Error("Failed to send scheduled notification", "error", err, "id", sched.ID.Hex())
		return
	}

	// Update last run time
	now := time.Now()
	sched.LastRunAt = &now
	if err := s.repo.Update(ctx, sched); err != nil {
		s.log.Error("Failed to update schedule", "error", err, "id", sched.ID.Hex())
	}

	s.log.Info("Successfully executed scheduled notification", "id", sched.ID.Hex())
}

// parseEmailRequest converts interface{} to SendEmailRequest
func (s *NotificationScheduler) parseEmailRequest(data interface{}) (*domain.SendEmailRequest, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var req domain.SendEmailRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

// parseSMSRequest converts interface{} to SendSMSRequest
func (s *NotificationScheduler) parseSMSRequest(data interface{}) (*domain.SendSMSRequest, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var req domain.SendSMSRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

// parseWebhookRequest converts interface{} to SendWebhookRequest
func (s *NotificationScheduler) parseWebhookRequest(data interface{}) (*domain.SendWebhookRequest, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var req domain.SendWebhookRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

// AddSchedule adds a new schedule
func (s *NotificationScheduler) AddSchedule(sched *domain.ScheduledNotification) error {
	ctx := context.Background()

	// Save to database
	if err := s.repo.Create(ctx, sched); err != nil {
		return err
	}

	// Register with cron if active
	if sched.IsActive {
		return s.registerSchedule(sched)
	}

	return nil
}

// RemoveSchedule removes a schedule
func (s *NotificationScheduler) RemoveSchedule(id string) error {
	// Remove from cron
	if entryID, exists := s.entries[id]; exists {
		s.cron.Remove(entryID)
		delete(s.entries, id)
	}

	// Delete from database
	return s.repo.Delete(context.Background(), id)
}
