package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScheduledNotification represents a scheduled notification
type ScheduledNotification struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID  string             `json:"tenant_id" bson:"tenantId"`
	Type      NotificationType   `json:"type" bson:"type"`         // email, sms, webhook
	Schedule  string             `json:"schedule" bson:"schedule"` // cron expression
	Request   interface{}        `json:"request" bson:"request"`
	NextRunAt time.Time          `json:"next_run_at" bson:"nextRunAt"`
	LastRunAt *time.Time         `json:"last_run_at,omitempty" bson:"lastRunAt,omitempty"`
	IsActive  bool               `json:"is_active" bson:"isActive"`
	Version   int                `json:"version" bson:"version"`
	CreatedAt time.Time          `json:"created_at" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updatedAt"`
	DeletedAt *time.Time         `json:"deleted_at,omitempty" bson:"deletedAt,omitempty"`
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID        string             `json:"tenant_id" bson:"tenantId"`
	UserID          string             `json:"user_id" bson:"userId"`
	EmailEnabled    bool               `json:"email_enabled" bson:"emailEnabled"`
	SMSEnabled      bool               `json:"sms_enabled" bson:"smsEnabled"`
	WebhookEnabled  bool               `json:"webhook_enabled" bson:"webhookEnabled"`
	EmailCategories map[string]bool    `json:"email_categories" bson:"emailCategories"` // marketing: false, alerts: true
	SMSCategories   map[string]bool    `json:"sms_categories" bson:"smsCategories"`
	QuietHoursStart string             `json:"quiet_hours_start" bson:"quietHoursStart"` // "22:00"
	QuietHoursEnd   string             `json:"quiet_hours_end" bson:"quietHoursEnd"`     // "08:00"
	Timezone        string             `json:"timezone" bson:"timezone"`
	Version         int                `json:"version" bson:"version"`
	CreatedAt       time.Time          `json:"created_at" bson:"createdAt"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updatedAt"`
	DeletedAt       *time.Time         `json:"deleted_at,omitempty" bson:"deletedAt,omitempty"`
}
