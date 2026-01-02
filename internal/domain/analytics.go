package domain

import "time"

// NotificationAnalytics represents aggregated analytics for notifications
type NotificationAnalytics struct {
	TenantID        string                         `json:"tenant_id" bson:"tenantId"`
	Period          string                         `json:"period" bson:"period"` // hourly, daily, weekly, monthly
	StartDate       time.Time                      `json:"start_date" bson:"startDate"`
	EndDate         time.Time                      `json:"end_date" bson:"endDate"`
	TotalSent       int64                          `json:"total_sent" bson:"totalSent"`
	TotalDelivered  int64                          `json:"total_delivered" bson:"totalDelivered"`
	TotalFailed     int64                          `json:"total_failed" bson:"totalFailed"`
	TotalBounced    int64                          `json:"total_bounced" bson:"totalBounced"`
	TotalRead       int64                          `json:"total_read" bson:"totalRead"`
	TotalClicked    int64                          `json:"total_clicked" bson:"totalClicked"`
	ByType          map[NotificationType]int64     `json:"by_type" bson:"byType"`
	ByPriority      map[NotificationPriority]int64 `json:"by_priority" bson:"byPriority"`
	ByStatus        map[NotificationStatus]int64   `json:"by_status" bson:"byStatus"`
	ByCategory      map[string]int64               `json:"by_category" bson:"byCategory"`
	DeliveryRate    float64                        `json:"delivery_rate" bson:"deliveryRate"`
	OpenRate        float64                        `json:"open_rate" bson:"openRate"`
	ClickRate       float64                        `json:"click_rate" bson:"clickRate"`
	BounceRate      float64                        `json:"bounce_rate" bson:"bounceRate"`
	AvgDeliveryTime float64                        `json:"avg_delivery_time" bson:"avgDeliveryTime"` // in seconds
}

// NotificationEvent represents a tracking event for a notification
type NotificationEvent struct {
	ID             string            `json:"id" bson:"_id,omitempty"`
	NotificationID string            `json:"notification_id" bson:"notificationId"`
	TenantID       string            `json:"tenant_id" bson:"tenantId"`
	EventType      string            `json:"event_type" bson:"eventType"` // sent, delivered, opened, clicked, bounced
	Timestamp      time.Time         `json:"timestamp" bson:"timestamp"`
	IPAddress      string            `json:"ip_address,omitempty" bson:"ipAddress,omitempty"`
	UserAgent      string            `json:"user_agent,omitempty" bson:"userAgent,omitempty"`
	LinkClicked    string            `json:"link_clicked,omitempty" bson:"linkClicked,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt      time.Time         `json:"created_at" bson:"createdAt"`
}

// DeliveryReport represents a delivery status report
type DeliveryReport struct {
	TenantID        string                 `json:"tenant_id"`
	Period          string                 `json:"period"`
	StartDate       time.Time              `json:"start_date"`
	EndDate         time.Time              `json:"end_date"`
	Summary         *NotificationAnalytics `json:"summary"`
	TopCategories   []CategoryStats        `json:"top_categories"`
	HourlyBreakdown []HourlyStats          `json:"hourly_breakdown"`
	FailureReasons  map[string]int64       `json:"failure_reasons"`
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
	Hour           int   `json:"hour"`
	TotalSent      int64 `json:"total_sent"`
	TotalDelivered int64 `json:"total_delivered"`
	TotalFailed    int64 `json:"total_failed"`
}
