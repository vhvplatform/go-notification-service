package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScheduledNotification represents a scheduled notification
type ScheduledNotification struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID   string             `json:"tenant_id" bson:"tenant_id"`
	Type       NotificationType   `json:"type" bson:"type"` // email, sms, webhook
	Schedule   string             `json:"schedule" bson:"schedule"` // cron expression
	Request    interface{}        `json:"request" bson:"request"`
	NextRunAt  time.Time          `json:"next_run_at" bson:"next_run_at"`
	LastRunAt  *time.Time         `json:"last_run_at,omitempty" bson:"last_run_at,omitempty"`
	IsActive   bool               `json:"is_active" bson:"is_active"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID        string             `json:"tenant_id" bson:"tenant_id"`
	UserID          string             `json:"user_id" bson:"user_id"`
	EmailEnabled    bool               `json:"email_enabled" bson:"email_enabled"`
	SMSEnabled      bool               `json:"sms_enabled" bson:"sms_enabled"`
	WebhookEnabled  bool               `json:"webhook_enabled" bson:"webhook_enabled"`
	EmailCategories map[string]bool    `json:"email_categories" bson:"email_categories"` // marketing: false, alerts: true
	SMSCategories   map[string]bool    `json:"sms_categories" bson:"sms_categories"`
	QuietHoursStart string             `json:"quiet_hours_start" bson:"quiet_hours_start"` // "22:00"
	QuietHoursEnd   string             `json:"quiet_hours_end" bson:"quiet_hours_end"`     // "08:00"
	Timezone        string             `json:"timezone" bson:"timezone"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}
