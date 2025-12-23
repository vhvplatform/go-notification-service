package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/longvhv/saas-framework-go/pkg/logger"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/domain"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/repository"
)

// EmailConfig holds email service configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// EmailService handles email operations
type EmailService struct {
	config       EmailConfig
	notifRepo    *repository.NotificationRepository
	templateRepo *repository.TemplateRepository
	log          *logger.Logger
}

// NewEmailService creates a new email service
func NewEmailService(config EmailConfig, notifRepo *repository.NotificationRepository, templateRepo *repository.TemplateRepository, log *logger.Logger) *EmailService {
	return &EmailService{
		config:       config,
		notifRepo:    notifRepo,
		templateRepo: templateRepo,
		log:          log,
	}
}

// SendEmail sends an email notification
func (s *EmailService) SendEmail(ctx context.Context, req *domain.SendEmailRequest) error {
	// Apply template if specified
	subject := req.Subject
	body := req.Body

	if req.TemplateID != "" {
		template, err := s.templateRepo.FindByID(ctx, req.TemplateID)
		if err != nil {
			s.log.Error("Failed to load template", "error", err, "template_id", req.TemplateID)
			return fmt.Errorf("failed to load template: %w", err)
		}

		subject = s.applyVariables(template.Subject, req.Variables)
		body = s.applyVariables(template.Body, req.Variables)
	}

	// Create notification record for each recipient
	for _, recipient := range req.To {
		notification := &domain.Notification{
			TenantID:  req.TenantID,
			Type:      domain.NotificationTypeEmail,
			Status:    domain.NotificationStatusPending,
			Recipient: recipient,
			Subject:   subject,
			Body:      body,
		}

		if err := s.notifRepo.Create(ctx, notification); err != nil {
			s.log.Error("Failed to create notification record", "error", err)
			continue
		}

		// Send email
		if err := s.sendSMTPEmail(recipient, subject, body, req.IsHTML); err != nil {
			s.log.Error("Failed to send email", "error", err, "recipient", recipient)
			s.notifRepo.UpdateStatus(ctx, notification.ID.Hex(), domain.NotificationStatusFailed, err.Error(), nil)
			continue
		}

		// Update status to sent
		now := ctx.Value("timestamp")
		if now == nil {
			t := context.Background().Value("timestamp")
			now = &t
		}
		s.notifRepo.UpdateStatus(ctx, notification.ID.Hex(), domain.NotificationStatusSent, "", nil)
	}

	return nil
}

// sendSMTPEmail sends an email via SMTP
func (s *EmailService) sendSMTPEmail(to, subject, body string, isHTML bool) error {
	from := s.config.FromEmail
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	}

	// Build email message
	var contentType string
	if isHTML {
		contentType = "text/html"
	} else {
		contentType = "text/plain"
	}

	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: %s; charset=UTF-8\r\n"+
		"\r\n"+
		"%s",
		from, to, subject, contentType, body)

	// Connect to SMTP server
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Use TLS if port is 465
	if s.config.SMTPPort == 465 {
		tlsConfig := &tls.Config{
			ServerName: s.config.SMTPHost,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.config.SMTPHost)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}

		if err = client.Mail(s.config.FromEmail); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to get data writer: %w", err)
		}

		_, err = w.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}

		return w.Close()
	}

	// Use STARTTLS for other ports
	return smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, []byte(message))
}

// applyVariables replaces template variables with actual values
func (s *EmailService) applyVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
