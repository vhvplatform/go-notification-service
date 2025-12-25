package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"html"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"github.com/vhvcorp/go-notification-service/internal/domain"
	"github.com/vhvcorp/go-notification-service/internal/repository"
	smtppool "github.com/vhvcorp/go-notification-service/internal/smtp"
	"github.com/vhvcorp/go-shared/logger"
)

// EmailConfig holds email service configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	PoolSize     int // Number of SMTP connections in the pool
}

// EmailService handles email operations
type EmailService struct {
	config       EmailConfig
	notifRepo    *repository.NotificationRepository
	templateRepo *repository.TemplateRepository
	log          *logger.Logger
	emailRegex   *regexp.Regexp
	smtpPool     *smtppool.SMTPPool
}

// NewEmailService creates a new email service
func NewEmailService(config EmailConfig, notifRepo *repository.NotificationRepository, templateRepo *repository.TemplateRepository, log *logger.Logger) *EmailService {
	// Compile email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Set default pool size if not specified
	poolSize := config.PoolSize
	if poolSize <= 0 {
		poolSize = 10
	}

	// Create SMTP pool
	smtpConfig := smtppool.SMTPConfig{
		Host:     config.SMTPHost,
		Port:     config.SMTPPort,
		Username: config.SMTPUsername,
		Password: config.SMTPPassword,
		UseTLS:   config.SMTPPort == 465,
	}

	smtpPool, err := smtppool.NewSMTPPool(smtpConfig, poolSize)
	if err != nil {
		log.Warn("Failed to create SMTP pool, will use direct connections", "error", err)
	}

	return &EmailService{
		config:       config,
		notifRepo:    notifRepo,
		templateRepo: templateRepo,
		log:          log,
		emailRegex:   emailRegex,
		smtpPool:     smtpPool,
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
		// Validate email address
		if !s.isValidEmail(recipient) {
			s.log.Warn("Invalid email address", "recipient", recipient)
			continue
		}

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

		// Update status to sent with current timestamp
		now := time.Now()
		s.notifRepo.UpdateStatus(ctx, notification.ID.Hex(), domain.NotificationStatusSent, "", &now)
	}

	return nil
}

// sendSMTPEmail sends an email via SMTP
func (s *EmailService) sendSMTPEmail(to, subject, body string, isHTML bool) error {
	// Try to use connection pool if available
	if s.smtpPool != nil {
		return s.sendViaSMTPPool(to, subject, body, isHTML)
	}

	// Fallback to direct connection
	return s.sendViaDirect(to, subject, body, isHTML)
}

// sendViaSMTPPool sends email using connection pool
func (s *EmailService) sendViaSMTPPool(to, subject, body string, isHTML bool) error {
	client, err := s.smtpPool.Get()
	if err != nil {
		s.log.Warn("Failed to get connection from pool, falling back to direct", "error", err)
		return s.sendViaDirect(to, subject, body, isHTML)
	}
	defer s.smtpPool.Put(client)

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

	// Send email using pooled connection
	if err := client.Mail(s.config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		w.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}

// sendViaDirect sends email using direct SMTP connection
func (s *EmailService) sendViaDirect(to, subject, body string, isHTML bool) error {
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
// Variables are HTML escaped to prevent XSS vulnerabilities
func (s *EmailService) applyVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		// HTML escape the value to prevent XSS
		escapedValue := html.EscapeString(value)
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, escapedValue)
	}
	return result
}

// isValidEmail validates email address format
func (s *EmailService) isValidEmail(email string) bool {
	return s.emailRegex.MatchString(email)
}

// Close closes the SMTP connection pool
func (s *EmailService) Close() {
	if s.smtpPool != nil {
		s.smtpPool.Close()
	}
}
