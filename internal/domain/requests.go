package domain

import (
	"time"
)

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	TenantID       string            `json:"tenant_id" binding:"required"`
	To             []string          `json:"to" binding:"required,min=1"`
	CC             []string          `json:"cc,omitempty"`
	BCC            []string          `json:"bcc,omitempty"`
	Subject        string            `json:"subject" binding:"required"`
	Body           string            `json:"body" binding:"required"`
	IsHTML         bool              `json:"is_html"`
	TemplateID     string            `json:"template_id,omitempty"`
	Variables      map[string]string `json:"variables,omitempty"`
	Attachments    []Attachment      `json:"attachments,omitempty"`
	Priority       NotificationPriority `json:"priority,omitempty"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Category       string            `json:"category,omitempty"`
	GroupID        string            `json:"group_id,omitempty"`
	ParentID       string            `json:"parent_id,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	ExpiresAt      *time.Time        `json:"expires_at,omitempty"`
	ScheduledFor   *time.Time        `json:"scheduled_for,omitempty"`
	TrackOpens     bool              `json:"track_opens,omitempty"`
	TrackClicks    bool              `json:"track_clicks,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string `json:"filename" binding:"required"`
	Content  []byte `json:"content" binding:"required"`
	MimeType string `json:"mime_type"`
}

// SendWebhookRequest represents a request to send a webhook
type SendWebhookRequest struct {
	TenantID       string            `json:"tenant_id" binding:"required"`
	URL            string            `json:"url" binding:"required,url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers,omitempty"`
	Payload        map[string]any    `json:"payload" binding:"required"`
	Timeout        int               `json:"timeout,omitempty"`
	Priority       NotificationPriority `json:"priority,omitempty"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Category       string            `json:"category,omitempty"`
	GroupID        string            `json:"group_id,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	RetryAttempts  int               `json:"retry_attempts,omitempty"`
}

// GetNotificationsRequest represents a request to get notifications
type GetNotificationsRequest struct {
	TenantID   string                   `form:"tenant_id" binding:"required"`
	Type       NotificationType         `form:"type"`
	Status     NotificationStatus       `form:"status"`
	Priority   NotificationPriority     `form:"priority"`
	Category   string                   `form:"category"`
	GroupID    string                   `form:"group_id"`
	Tags       []string                 `form:"tags"`
	Page       int                      `form:"page"`
	PageSize   int                      `form:"page_size"`
}

// SendSMSRequest represents a request to send an SMS
type SendSMSRequest struct {
	TenantID       string            `json:"tenant_id" binding:"required"`
	To             string            `json:"to" binding:"required"`
	Message        string            `json:"message" binding:"required"`
	Priority       NotificationPriority `json:"priority,omitempty"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Category       string            `json:"category,omitempty"`
	GroupID        string            `json:"group_id,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	ScheduledFor   *time.Time        `json:"scheduled_for,omitempty"`
}

// BulkEmailRequest represents a request to send bulk emails
type BulkEmailRequest struct {
	TenantID       string            `json:"tenant_id" binding:"required"`
	Recipients     []string          `json:"recipients" binding:"required,min=1"`
	Subject        string            `json:"subject" binding:"required"`
	Body           string            `json:"body" binding:"required"`
	IsHTML         bool              `json:"is_html"`
	TemplateID     string            `json:"template_id,omitempty"`
	Variables      map[string]string `json:"variables,omitempty"`
	Priority       int               `json:"priority"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Category       string            `json:"category,omitempty"`
	GroupID        string            `json:"group_id,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	TrackOpens     bool              `json:"track_opens,omitempty"`
	TrackClicks    bool              `json:"track_clicks,omitempty"`
}

// NotificationStatusUpdate represents a status update for a notification
type NotificationStatusUpdate struct {
	NotificationID string             `json:"notification_id" binding:"required"`
	Status         NotificationStatus `json:"status" binding:"required"`
	Timestamp      time.Time          `json:"timestamp"`
	IPAddress      string             `json:"ip_address,omitempty"`
	UserAgent      string             `json:"user_agent,omitempty"`
	LinkClicked    string             `json:"link_clicked,omitempty"`
}

// NotificationSearchRequest represents advanced search criteria
type NotificationSearchRequest struct {
	TenantID      string                   `form:"tenant_id" binding:"required"`
	Type          NotificationType         `form:"type"`
	Status        NotificationStatus       `form:"status"`
	Priority      NotificationPriority     `form:"priority"`
	Category      string                   `form:"category"`
	GroupID       string                   `form:"group_id"`
	Tags          []string                 `form:"tags"`
	Recipient     string                   `form:"recipient"`
	Subject       string                   `form:"subject"`
	FromDate      *time.Time               `form:"from_date"`
	ToDate        *time.Time               `form:"to_date"`
	Page          int                      `form:"page"`
	PageSize      int                      `form:"page_size"`
	SortBy        string                   `form:"sort_by"` // created_at, sent_at, priority
	SortOrder     string                   `form:"sort_order"` // asc, desc
}
