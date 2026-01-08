package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// NotificationsSent tracks the total number of notifications sent
	NotificationsSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_service_sent_total",
			Help: "Total number of notifications sent",
		},
		[]string{"type", "tenant_id", "status"},
	)

	// NotificationDuration tracks notification sending duration
	NotificationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "notification_service_duration_seconds",
			Help:    "Notification sending duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	// EmailQueueSize tracks the current email queue size
	EmailQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "notification_service_email_queue_size",
			Help: "Current number of emails in the priority queue",
		},
	)

	// SMTPConnectionPool tracks the number of SMTP connections in the pool
	SMTPConnectionPool = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "notification_service_smtp_connections",
			Help: "Number of active SMTP connections in the pool",
		},
	)

	// FailedNotifications tracks the number of failed notifications
	FailedNotifications = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_service_failed_total",
			Help: "Total number of failed notifications",
		},
		[]string{"type", "tenant_id", "reason"},
	)

	// DLQSize tracks the size of the dead letter queue
	DLQSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "notification_service_dlq_size",
			Help: "Number of notifications in the dead letter queue",
		},
	)

	// EmailBounces tracks email bounce events
	EmailBounces = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_service_email_bounces_total",
			Help: "Total number of email bounce events",
		},
		[]string{"type"}, // hard, soft, complaint
	)

	// RateLimitExceeded tracks rate limit violations
	RateLimitExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_service_rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded events",
		},
		[]string{"tenant_id"},
	)

	// ConsumerRestarts tracks event consumer restart events
	ConsumerRestarts = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "notification_service_consumer_restarts_total",
			Help: "Total number of event consumer restarts",
		},
	)
)
