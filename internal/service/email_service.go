package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html"
	"net/smtp"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/repository"
	"github.com/vhvplatform/go-notification-service/internal/shared/logger"
	smtppool "github.com/vhvplatform/go-notification-service/internal/smtp"
)

// Security constants
const (
	maxEmailLength     = 320  // Maximum email address length per RFC 5321
	maxSubjectLength   = 998  // Maximum subject line length per RFC 5322
	maxBodyLength      = 10 * 1024 * 1024 // Maximum email body size: 10MB
	maxRecipientsCount = 1000 // Maximum recipients per email
	maxVariableKeyLen  = 256  // Maximum variable key length
	maxVariableValLen  = 65536 // Maximum variable value length: 64KB
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

// validateEmailInput performs security validation on email input
func validateEmailInput(to []string, subject, body string, variables map[string]string) error {
	// Validate recipients
	if len(to) == 0 {
		return errors.New("at least one recipient is required")
	}
	if len(to) > maxRecipientsCount {
		return fmt.Errorf("too many recipients: %d (max: %d)", len(to), maxRecipientsCount)
	}

	// Validate subject length
	if len(subject) > maxSubjectLength {
		return fmt.Errorf("subject too long: %d bytes (max: %d)", len(subject), maxSubjectLength)
	}

	// Validate body length
	if len(body) > maxBodyLength {
		return fmt.Errorf("body too large: %d bytes (max: %d)", len(body), maxBodyLength)
	}

	// Validate UTF-8 encoding
	if !utf8.ValidString(subject) {
		return errors.New("subject contains invalid UTF-8 characters")
	}
	if !utf8.ValidString(body) {
		return errors.New("body contains invalid UTF-8 characters")
	}

	// Validate variables
	for key, value := range variables {
		if len(key) > maxVariableKeyLen {
			return fmt.Errorf("variable key too long: %d bytes (max: %d)", len(key), maxVariableKeyLen)
		}
		if len(value) > maxVariableValLen {
			return fmt.Errorf("variable value too long: %d bytes (max: %d)", len(value), maxVariableValLen)
		}
		if !utf8.ValidString(key) || !utf8.ValidString(value) {
			return errors.New("variable contains invalid UTF-8 characters")
		}
		// Prevent null bytes in variables
		if strings.Contains(key, "\x00") || strings.Contains(value, "\x00") {
			return errors.New("variable contains null bytes")
		}
	}

	return nil
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

// SendEmail sends an email notification with optimized batch processing and security validation
func (s *EmailService) SendEmail(ctx context.Context, req *domain.SendEmailRequest) error {
	// Check idempotency key if provided
	if req.IdempotencyKey != "" {
		existing, err := s.notifRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && existing != nil {
			s.log.Info("Idempotent request - notification already exists", "idempotency_key", req.IdempotencyKey, "notification_id", existing.ID.Hex())
			return nil // Already processed
		}
	}

	// Security validation on input
	if err := validateEmailInput(req.To, req.Subject, req.Body, req.Variables); err != nil {
		s.log.Warn("Email input validation failed", "error", err)
		return fmt.Errorf("invalid email input: %w", err)
	}

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

	// Set default priority if not specified
	priority := req.Priority
	if priority == "" {
		priority = domain.NotificationPriorityNormal
	}

	// Validate recipients and create notification records in batch
	var validRecipients []string
	var notifications []*domain.Notification

	for _, recipient := range req.To {
		// Validate email address
		if !s.isValidEmail(recipient) {
			s.log.Warn("Invalid email address", "recipient", recipient)
			continue
		}

		validRecipients = append(validRecipients, recipient)
		notifications = append(notifications, &domain.Notification{
			TenantID:       req.TenantID,
			Type:           domain.NotificationTypeEmail,
			Status:         domain.NotificationStatusPending,
			Priority:       priority,
			Recipient:      recipient,
			Subject:        subject,
			Body:           body,
			IdempotencyKey: req.IdempotencyKey,
			Tags:           req.Tags,
			Category:       req.Category,
			GroupID:        req.GroupID,
			ParentID:       req.ParentID,
			Metadata:       req.Metadata,
			ExpiresAt:      req.ExpiresAt,
			ScheduledFor:   req.ScheduledFor,
		})
	}

	if len(notifications) == 0 {
		return fmt.Errorf("no valid recipients")
	}

	// Batch create notification records
	if err := s.notifRepo.CreateBatch(ctx, notifications); err != nil {
		s.log.Error("Failed to create notification records", "error", err)
		return err
	}

	// Send emails
	for i, recipient := range validRecipients {
		if err := s.sendSMTPEmail(recipient, subject, body, req.IsHTML); err != nil {
			s.log.Error("Failed to send email", "error", err, "recipient", recipient)
			s.notifRepo.UpdateStatus(ctx, notifications[i].ID.Hex(), domain.NotificationStatusFailed, err.Error(), nil)
			continue
		}

		// Update status to sent with current timestamp
		now := time.Now()
		s.notifRepo.UpdateStatus(ctx, notifications[i].ID.Hex(), domain.NotificationStatusSent, "", &now)
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
// Uses strings.Replacer for efficient multiple replacements
func (s *EmailService) applyVariables(template string, variables map[string]string) string {
	if len(variables) == 0 {
		return template
	}

	// Build replacement pairs for strings.Replacer (more efficient than multiple ReplaceAll)
	replacements := make([]string, 0, len(variables)*2)
	for key, value := range variables {
		// HTML escape the value to prevent XSS
		escapedValue := html.EscapeString(value)
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacements = append(replacements, placeholder, escapedValue)
	}

	// Use strings.Replacer for efficient batch replacement
	replacer := strings.NewReplacer(replacements...)
	return replacer.Replace(template)
}

// isValidEmail validates email address format with security checks
func (s *EmailService) isValidEmail(email string) bool {
	// Basic security checks
	if len(email) == 0 || len(email) > maxEmailLength {
		return false
	}
	
	// Check for null bytes and control characters
	if strings.ContainsAny(email, "\x00\r\n") {
		return false
	}
	
	// Validate UTF-8 encoding
	if !utf8.ValidString(email) {
		return false
	}
	
	// Apply regex validation
	return s.emailRegex.MatchString(email)
}

// Close closes the SMTP connection pool
func (s *EmailService) Close() {
	if s.smtpPool != nil {
		s.smtpPool.Close()
	}
}
