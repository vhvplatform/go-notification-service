package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail   NotificationType = "email"
	NotificationTypeWebhook NotificationType = "webhook"
	NotificationTypeSMS     NotificationType = "sms"
)

// NotificationPriority represents the priority level of a notification
type NotificationPriority string

const (
	NotificationPriorityCritical NotificationPriority = "critical" // Immediate delivery, bypass rate limits
	NotificationPriorityHigh     NotificationPriority = "high"     // High priority, fast processing
	NotificationPriorityNormal   NotificationPriority = "normal"   // Standard priority
	NotificationPriorityLow      NotificationPriority = "low"      // Low priority, can be delayed
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusQueued    NotificationStatus = "queued"    // Queued for processing
	NotificationStatusSending   NotificationStatus = "sending"   // Currently being sent
	NotificationStatusSent      NotificationStatus = "sent"      // Successfully sent to provider
	NotificationStatusDelivered NotificationStatus = "delivered" // Confirmed delivered to recipient
	NotificationStatusFailed    NotificationStatus = "failed"    // Failed to send
	NotificationStatusBounced   NotificationStatus = "bounced"   // Email bounced
	NotificationStatusRead      NotificationStatus = "read"      // Recipient opened/read the notification
	NotificationStatusClicked   NotificationStatus = "clicked"   // Recipient clicked links in notification
)

// Notification represents a notification record
type Notification struct {
	ID              primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	TenantID        string               `json:"tenant_id" bson:"tenantId"`
	Type            NotificationType     `json:"type" bson:"type"`
	Status          NotificationStatus   `json:"status" bson:"status"`
	Priority        NotificationPriority `json:"priority" bson:"priority"`
	Recipient       string               `json:"recipient" bson:"recipient"`
	Subject         string               `json:"subject,omitempty" bson:"subject,omitempty"`
	Body            string               `json:"body,omitempty" bson:"body,omitempty"`
	Payload         map[string]any       `json:"payload,omitempty" bson:"payload,omitempty"`
	Error           string               `json:"error,omitempty" bson:"error,omitempty"`
	RetryCount      int                  `json:"retry_count" bson:"retryCount"`
	IdempotencyKey  string               `json:"idempotency_key,omitempty" bson:"idempotencyKey,omitempty"`
	Tags            []string             `json:"tags,omitempty" bson:"tags,omitempty"`
	Category        string               `json:"category,omitempty" bson:"category,omitempty"`
	GroupID         string               `json:"group_id,omitempty" bson:"groupId,omitempty"`
	ParentID        string               `json:"parent_id,omitempty" bson:"parentId,omitempty"`
	Metadata        map[string]string    `json:"metadata,omitempty" bson:"metadata,omitempty"`
	SentAt          *time.Time           `json:"sent_at,omitempty" bson:"sentAt,omitempty"`
	DeliveredAt     *time.Time           `json:"delivered_at,omitempty" bson:"deliveredAt,omitempty"`
	ReadAt          *time.Time           `json:"read_at,omitempty" bson:"readAt,omitempty"`
	ClickedAt       *time.Time           `json:"clicked_at,omitempty" bson:"clickedAt,omitempty"`
	ExpiresAt       *time.Time           `json:"expires_at,omitempty" bson:"expiresAt,omitempty"`
	ScheduledFor    *time.Time           `json:"scheduled_for,omitempty" bson:"scheduledFor,omitempty"`
	CreatedAt       time.Time            `json:"created_at" bson:"createdAt"`
	UpdatedAt       time.Time            `json:"updated_at" bson:"updatedAt"`
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID  string             `json:"tenant_id" bson:"tenantId"`
	Name      string             `json:"name" bson:"name"`
	Subject   string             `json:"subject" bson:"subject"`
	Body      string             `json:"body" bson:"body"`
	IsHTML    bool               `json:"is_html" bson:"isHtml"`
	Variables []string           `json:"variables,omitempty" bson:"variables,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updatedAt"`
}

// EventType represents the type of event
type EventType string

const (
	EventUserRegistered    EventType = "user.registered"
	EventUserPasswordReset EventType = "user.password_reset"
	EventTenantCreated     EventType = "tenant.created"
	EventPaymentCompleted  EventType = "payment.completed"
)

// Event represents an event from RabbitMQ
type Event struct {
	Type      EventType      `json:"type"`
	TenantID  string         `json:"tenant_id"`
	UserID    string         `json:"user_id,omitempty"`
	Email     string         `json:"email,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// FailedNotification represents a notification that failed after all retries
type FailedNotification struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OriginalID primitive.ObjectID `json:"original_id" bson:"originalId"`
	TenantID   string             `json:"tenant_id" bson:"tenantId"`
	Type       NotificationType   `json:"type" bson:"type"`
	Recipient  string             `json:"recipient" bson:"recipient"`
	Subject    string             `json:"subject,omitempty" bson:"subject,omitempty"`
	Body       string             `json:"body,omitempty" bson:"body,omitempty"`
	Payload    map[string]any     `json:"payload,omitempty" bson:"payload,omitempty"`
	Error      string             `json:"error" bson:"error"`
	FailedAt   time.Time          `json:"failed_at" bson:"failedAt"`
	RetryCount int                `json:"retry_count" bson:"retryCount"`
	CreatedAt  time.Time          `json:"created_at" bson:"createdAt"`
}

// EmailBounce represents an email bounce record
type EmailBounce struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email"`
	Type      string             `json:"type" bson:"type"` // hard, soft, complaint
	Reason    string             `json:"reason" bson:"reason"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	CreatedAt time.Time          `json:"created_at" bson:"createdAt"`
}
