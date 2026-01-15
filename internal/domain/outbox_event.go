package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OutboxEventStatus represents the processing status of an outbox event
type OutboxEventStatus string

const (
	OutboxEventStatusPending   OutboxEventStatus = "pending"
	OutboxEventStatusProcessed OutboxEventStatus = "processed"
	OutboxEventStatusFailed    OutboxEventStatus = "failed"
)

// OutboxEventType represents the type of domain event
type OutboxEventType string

const (
	// Notification Events
	EventNotificationCreated       OutboxEventType = "notification.created"
	EventNotificationStatusChanged OutboxEventType = "notification.status_changed"
	EventNotificationUpdated       OutboxEventType = "notification.updated"
	EventNotificationDeleted       OutboxEventType = "notification.deleted"

	// Template Events
	EventTemplateCreated OutboxEventType = "template.created"
	EventTemplateUpdated OutboxEventType = "template.updated"
	EventTemplateDeleted OutboxEventType = "template.deleted"

	// Scheduled Notification Events
	EventScheduledNotificationCreated  OutboxEventType = "scheduled_notification.created"
	EventScheduledNotificationExecuted OutboxEventType = "scheduled_notification.executed"
	EventScheduledNotificationCanceled OutboxEventType = "scheduled_notification.canceled"
)

// OutboxEvent represents an event to be published via Debezium CDC
// This table is the ONLY table that Debezium should CDC from (per Architecture Rules)
type OutboxEvent struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// Standard Fields (Phase 1)
	TenantID  string     `bson:"tenantId" json:"tenantId"`
	Version   int        `bson:"version" json:"version"`
	CreatedAt *time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt *time.Time `bson:"updatedAt" json:"updatedAt"`
	DeletedAt *time.Time `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`

	// Aggregate Information
	AggregateType string `bson:"aggregateType" json:"aggregateType"` // "notification", "template", etc.
	AggregateID   string `bson:"aggregateId" json:"aggregateId"`     // ID of the changed entity

	// Event Details
	EventType OutboxEventType `bson:"eventType" json:"eventType"` // "notification.created", etc.
	Payload   interface{}     `bson:"payload" json:"payload"`     // JSON payload with event data

	// Distributed Tracing (Phase 2 - CRITICAL for OpenTelemetry)
	TraceID string `bson:"traceId" json:"traceId"` // OpenTelemetry trace ID
	SpanID  string `bson:"spanId" json:"spanId"`   // OpenTelemetry span ID

	// Processing Status
	Status      OutboxEventStatus `bson:"status" json:"status"`
	ProcessedAt *time.Time        `bson:"processedAt,omitempty" json:"processedAt,omitempty"`
	ErrorCount  int               `bson:"errorCount" json:"errorCount"`         // Retry count for failed events
	LastError   string            `bson:"lastError,omitempty" json:"lastError"` // Last error message
}

// NotificationCreatedPayload represents the payload for notification.created event
type NotificationCreatedPayload struct {
	NotificationID string             `json:"notificationId"`
	TenantID       string             `json:"tenantId"`
	Type           NotificationType   `json:"type"`
	Recipient      string             `json:"recipient"`
	Subject        string             `json:"subject,omitempty"`
	Status         NotificationStatus `json:"status"`
	CreatedAt      time.Time          `json:"createdAt"`
}

// NotificationStatusChangedPayload represents the payload for notification.status_changed event
type NotificationStatusChangedPayload struct {
	NotificationID string             `json:"notificationId"`
	TenantID       string             `json:"tenantId"`
	OldStatus      NotificationStatus `json:"oldStatus"`
	NewStatus      NotificationStatus `json:"newStatus"`
	ChangedAt      time.Time          `json:"changedAt"`
}

// NotificationUpdatedPayload represents the payload for notification.updated event
type NotificationUpdatedPayload struct {
	NotificationID string           `json:"notificationId"`
	TenantID       string           `json:"tenantId"`
	Type           NotificationType `json:"type"`
	UpdatedFields  []string         `json:"updatedFields"` // List of changed fields
	UpdatedAt      time.Time        `json:"updatedAt"`
}

// NotificationDeletedPayload represents the payload for notification.deleted event
type NotificationDeletedPayload struct {
	NotificationID string    `json:"notificationId"`
	TenantID       string    `json:"tenantId"`
	DeletedAt      time.Time `json:"deletedAt"`
}

// TemplateCreatedPayload represents the payload for template.created event
type TemplateCreatedPayload struct {
	TemplateID   string    `json:"templateId"`
	TenantID     string    `json:"tenantId"`
	Name         string    `json:"name"`
	TemplateType string    `json:"templateType"`
	CreatedAt    time.Time `json:"createdAt"`
}

// TemplateUpdatedPayload represents the payload for template.updated event
type TemplateUpdatedPayload struct {
	TemplateID    string    `json:"templateId"`
	TenantID      string    `json:"tenantId"`
	Name          string    `json:"name"`
	UpdatedFields []string  `json:"updatedFields"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// TemplateDeletedPayload represents the payload for template.deleted event
type TemplateDeletedPayload struct {
	TemplateID string    `json:"templateId"`
	TenantID   string    `json:"tenantId"`
	Name       string    `json:"name"`
	DeletedAt  time.Time `json:"deletedAt"`
}

// ScheduledNotificationCreatedPayload represents the payload for scheduled_notification.created event
type ScheduledNotificationCreatedPayload struct {
	ScheduleID     string    `json:"scheduleId"`
	TenantID       string    `json:"tenantId"`
	NotificationID string    `json:"notificationId"`
	ScheduledFor   time.Time `json:"scheduledFor"`
	CreatedAt      time.Time `json:"createdAt"`
}

// ScheduledNotificationExecutedPayload represents the payload for scheduled_notification.executed event
type ScheduledNotificationExecutedPayload struct {
	ScheduleID     string    `json:"scheduleId"`
	TenantID       string    `json:"tenantId"`
	NotificationID string    `json:"notificationId"`
	ExecutedAt     time.Time `json:"executedAt"`
}

// ScheduledNotificationCanceledPayload represents the payload for scheduled_notification.canceled event
type ScheduledNotificationCanceledPayload struct {
	ScheduleID string    `json:"scheduleId"`
	TenantID   string    `json:"tenantId"`
	CanceledAt time.Time `json:"canceledAt"`
}
