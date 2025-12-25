package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vhvcorp/go-notification-service/internal/domain"
	"github.com/vhvcorp/go-notification-service/internal/repository"
	"github.com/vhvcorp/go-notification-service/internal/shared/logger"
)

// SMSConfig holds SMS service configuration
type SMSConfig struct {
	Provider    string // twilio, aws_sns
	TwilioSID   string
	TwilioToken string
	TwilioFrom  string
	AWSSNSARN   string
	AWSRegion   string
}

// SMSService handles SMS operations
type SMSService struct {
	config    SMSConfig
	notifRepo *repository.NotificationRepository
	log       *logger.Logger
}

// NewSMSService creates a new SMS service
func NewSMSService(config SMSConfig, notifRepo *repository.NotificationRepository, log *logger.Logger) *SMSService {
	return &SMSService{
		config:    config,
		notifRepo: notifRepo,
		log:       log,
	}
}

// SendSMS sends an SMS notification
func (s *SMSService) SendSMS(ctx context.Context, req *domain.SendSMSRequest) error {
	// Create notification record
	notification := &domain.Notification{
		TenantID:  req.TenantID,
		Type:      domain.NotificationTypeSMS,
		Status:    domain.NotificationStatusPending,
		Recipient: req.To,
		Body:      req.Message,
	}

	if err := s.notifRepo.Create(ctx, notification); err != nil {
		s.log.Error("Failed to create notification record", "error", err)
		return err
	}

	// Send SMS based on provider
	var err error
	switch s.config.Provider {
	case "twilio":
		err = s.sendViaTwilio(req)
	case "aws_sns":
		err = s.sendViaAWSSNS(req)
	default:
		err = fmt.Errorf("unsupported SMS provider: %s", s.config.Provider)
	}

	if err != nil {
		s.log.Error("Failed to send SMS", "error", err)
		now := time.Now()
		s.notifRepo.UpdateStatus(ctx, notification.ID.Hex(), domain.NotificationStatusFailed, err.Error(), &now)
		return err
	}

	// Update status
	now := time.Now()
	s.notifRepo.UpdateStatus(ctx, notification.ID.Hex(), domain.NotificationStatusSent, "", &now)
	return nil
}

// sendViaTwilio sends SMS via Twilio
func (s *SMSService) sendViaTwilio(req *domain.SendSMSRequest) error {
	// Note: Actual Twilio integration would require the twilio-go SDK
	// For now, this is a placeholder that logs the attempt
	s.log.Info("Sending SMS via Twilio", "to", req.To, "provider", "twilio")

	// TODO: Implement actual Twilio integration when SDK is added
	// This requires adding: github.com/twilio/twilio-go to dependencies
	/*
		client := twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: s.config.TwilioSID,
			Password: s.config.TwilioToken,
		})

		params := &twilioApi.CreateMessageParams{}
		params.SetTo(req.To)
		params.SetFrom(s.config.TwilioFrom)
		params.SetBody(req.Message)

		_, err := client.Api.CreateMessage(params)
		return err
	*/

	return nil
}

// sendViaAWSSNS sends SMS via AWS SNS
func (s *SMSService) sendViaAWSSNS(req *domain.SendSMSRequest) error {
	// Note: Actual AWS SNS integration would require AWS SDK
	// For now, this is a placeholder
	s.log.Info("Sending SMS via AWS SNS", "to", req.To, "provider", "aws_sns")

	// TODO: Implement AWS SNS integration when SDK is needed
	return fmt.Errorf("AWS SNS not implemented yet")
}
