package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vhvcorp/go-shared/logger"
	"github.com/vhvcorp/go-notification-service/internal/domain"
	"github.com/vhvcorp/go-notification-service/internal/repository"
)

// WebhookService handles webhook operations
type WebhookService struct {
	notifRepo *repository.NotificationRepository
	log       *logger.Logger
	client    *http.Client
}

// NewWebhookService creates a new webhook service
func NewWebhookService(notifRepo *repository.NotificationRepository, log *logger.Logger) *WebhookService {
	return &WebhookService{
		notifRepo: notifRepo,
		log:       log,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendWebhook sends a webhook notification
func (s *WebhookService) SendWebhook(ctx context.Context, req *domain.SendWebhookRequest) error {
	// Create notification record
	notification := &domain.Notification{
		TenantID:  req.TenantID,
		Type:      domain.NotificationTypeWebhook,
		Status:    domain.NotificationStatusPending,
		Recipient: req.URL,
		Payload:   req.Payload,
	}

	if err := s.notifRepo.Create(ctx, notification); err != nil {
		s.log.Error("Failed to create notification record", "error", err)
		return err
	}

	// Send webhook with retry logic
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// Exponential backoff
			backoff := time.Duration(i*i) * time.Second
			s.log.Info("Retrying webhook", "attempt", i+1, "backoff", backoff)
			time.Sleep(backoff)
			s.notifRepo.IncrementRetryCount(ctx, notification.ID.Hex())
		}

		if err := s.sendHTTPRequest(req); err != nil {
			lastErr = err
			s.log.Error("Failed to send webhook", "error", err, "attempt", i+1, "url", req.URL)
			continue
		}

		// Success
		now := time.Now()
		s.notifRepo.UpdateStatus(ctx, notification.ID.Hex(), domain.NotificationStatusSent, "", &now)
		return nil
	}

	// All retries failed
	s.notifRepo.UpdateStatus(ctx, notification.ID.Hex(), domain.NotificationStatusFailed, lastErr.Error(), nil)
	return fmt.Errorf("webhook failed after %d attempts: %w", maxRetries, lastErr)
}

// sendHTTPRequest sends an HTTP request to the webhook URL
func (s *WebhookService) sendHTTPRequest(req *domain.SendWebhookRequest) error {
	method := req.Method
	if method == "" {
		method = "POST"
	}

	timeout := req.Timeout
	if timeout == 0 {
		timeout = 30
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(method, req.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "SaaS-Framework-Notification-Service/1.0")
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Send request with custom timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
