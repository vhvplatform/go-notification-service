package domain

import "time"

// NotificationAnalytics represents aggregated analytics for notifications
type NotificationAnalytics struct {
	TenantID      string                        `json:"tenant_id" bson:"tenant_id"`
	Period        string                        `json:"period" bson:"period"` // hourly, daily, weekly, monthly
	StartDate     time.Time                     `json:"start_date" bson:"start_date"`
	EndDate       time.Time                     `json:"end_date" bson:"end_date"`
	TotalSent     int64                         `json:"total_sent" bson:"total_sent"`
	TotalDelivered int64                        `json:"total_delivered" bson:"total_delivered"`
	TotalFailed   int64                         `json:"total_failed" bson:"total_failed"`
	TotalBounced  int64                         `json:"total_bounced" bson:"total_bounced"`
	TotalRead     int64                         `json:"total_read" bson:"total_read"`
	TotalClicked  int64                         `json:"total_clicked" bson:"total_clicked"`
	ByType        map[NotificationType]int64    `json:"by_type" bson:"by_type"`
	ByPriority    map[NotificationPriority]int64 `json:"by_priority" bson:"by_priority"`
	ByStatus      map[NotificationStatus]int64  `json:"by_status" bson:"by_status"`
	ByCategory    map[string]int64              `json:"by_category" bson:"by_category"`
	DeliveryRate  float64                       `json:"delivery_rate" bson:"delivery_rate"`
	OpenRate      float64                       `json:"open_rate" bson:"open_rate"`
	ClickRate     float64                       `json:"click_rate" bson:"click_rate"`
	BounceRate    float64                       `json:"bounce_rate" bson:"bounce_rate"`
	AvgDeliveryTime float64                     `json:"avg_delivery_time" bson:"avg_delivery_time"` // in seconds
}

// NotificationEvent represents a tracking event for a notification
type NotificationEvent struct {
	ID             string             `json:"id" bson:"_id,omitempty"`
	NotificationID string             `json:"notification_id" bson:"notification_id"`
	TenantID       string             `json:"tenant_id" bson:"tenant_id"`
	EventType      string             `json:"event_type" bson:"event_type"` // sent, delivered, opened, clicked, bounced
	Timestamp      time.Time          `json:"timestamp" bson:"timestamp"`
	IPAddress      string             `json:"ip_address,omitempty" bson:"ip_address,omitempty"`
	UserAgent      string             `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
	LinkClicked    string             `json:"link_clicked,omitempty" bson:"link_clicked,omitempty"`
	Metadata       map[string]string  `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
}

// DeliveryReport represents a delivery status report
type DeliveryReport struct {
	TenantID       string                       `json:"tenant_id"`
	Period         string                       `json:"period"`
	StartDate      time.Time                    `json:"start_date"`
	EndDate        time.Time                    `json:"end_date"`
	Summary        *NotificationAnalytics       `json:"summary"`
	TopCategories  []CategoryStats              `json:"top_categories"`
	HourlyBreakdown []HourlyStats               `json:"hourly_breakdown"`
	FailureReasons map[string]int64             `json:"failure_reasons"`
}

// CategoryStats represents statistics for a category
type CategoryStats struct {
	Category     string  `json:"category"`
	TotalSent    int64   `json:"total_sent"`
	DeliveryRate float64 `json:"delivery_rate"`
	OpenRate     float64 `json:"open_rate"`
}

// HourlyStats represents hourly statistics
type HourlyStats struct {
	Hour         int   `json:"hour"`
	TotalSent    int64 `json:"total_sent"`
	TotalDelivered int64 `json:"total_delivered"`
	TotalFailed  int64 `json:"total_failed"`
}
