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

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusDelivered NotificationStatus = "delivered"
)

// Notification represents a notification record
type Notification struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID   string             `json:"tenant_id" bson:"tenant_id"`
	Type       NotificationType   `json:"type" bson:"type"`
	Status     NotificationStatus `json:"status" bson:"status"`
	Recipient  string             `json:"recipient" bson:"recipient"`
	Subject    string             `json:"subject,omitempty" bson:"subject,omitempty"`
	Body       string             `json:"body,omitempty" bson:"body,omitempty"`
	Payload    map[string]any     `json:"payload,omitempty" bson:"payload,omitempty"`
	Error      string             `json:"error,omitempty" bson:"error,omitempty"`
	RetryCount int                `json:"retry_count" bson:"retry_count"`
	SentAt     *time.Time         `json:"sent_at,omitempty" bson:"sent_at,omitempty"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID  string             `json:"tenant_id" bson:"tenant_id"`
	Name      string             `json:"name" bson:"name"`
	Subject   string             `json:"subject" bson:"subject"`
	Body      string             `json:"body" bson:"body"`
	IsHTML    bool               `json:"is_html" bson:"is_html"`
	Variables []string           `json:"variables,omitempty" bson:"variables,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
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
