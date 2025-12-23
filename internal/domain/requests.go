package domain

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	TenantID    string            `json:"tenant_id" binding:"required"`
	To          []string          `json:"to" binding:"required,min=1"`
	CC          []string          `json:"cc,omitempty"`
	BCC         []string          `json:"bcc,omitempty"`
	Subject     string            `json:"subject" binding:"required"`
	Body        string            `json:"body" binding:"required"`
	IsHTML      bool              `json:"is_html"`
	TemplateID  string            `json:"template_id,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string `json:"filename" binding:"required"`
	Content  []byte `json:"content" binding:"required"`
	MimeType string `json:"mime_type"`
}

// SendWebhookRequest represents a request to send a webhook
type SendWebhookRequest struct {
	TenantID string            `json:"tenant_id" binding:"required"`
	URL      string            `json:"url" binding:"required,url"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers,omitempty"`
	Payload  map[string]any    `json:"payload" binding:"required"`
	Timeout  int               `json:"timeout,omitempty"`
}

// GetNotificationsRequest represents a request to get notifications
type GetNotificationsRequest struct {
	TenantID string             `form:"tenant_id" binding:"required"`
	Type     NotificationType   `form:"type"`
	Status   NotificationStatus `form:"status"`
	Page     int                `form:"page"`
	PageSize int                `form:"page_size"`
}
