package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/longvhv/saas-shared-go/logger"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/domain"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/metrics"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/queue"
)

// BulkEmailService handles bulk email operations
type BulkEmailService struct {
	emailService *EmailService
	queue        *queue.PriorityQueue
	workers      int
	log          *logger.Logger
	stopChan     chan struct{}
}

// NewBulkEmailService creates a new bulk email service
func NewBulkEmailService(emailService *EmailService, workers int, log *logger.Logger) *BulkEmailService {
	if workers <= 0 {
		workers = 5 // Default to 5 workers
	}

	return &BulkEmailService{
		emailService: emailService,
		queue:        queue.NewPriorityQueue(),
		workers:      workers,
		log:          log,
		stopChan:     make(chan struct{}),
	}
}

// Start starts the worker pool
func (s *BulkEmailService) Start() {
	s.log.Info("Starting bulk email service", "workers", s.workers)

	for i := 0; i < s.workers; i++ {
		go s.worker(i)
	}
}

// Stop stops the worker pool
func (s *BulkEmailService) Stop() {
	close(s.stopChan)
}

// worker processes jobs from the queue
func (s *BulkEmailService) worker(id int) {
	s.log.Info("Starting bulk email worker", "worker_id", id)

	for {
		select {
		case <-s.stopChan:
			s.log.Info("Stopping bulk email worker", "worker_id", id)
			return
		default:
			// Use blocking Pop instead of TryPop with sleep
			job := s.queue.Pop() // This blocks until a job is available

			// Update queue size metric
			metrics.EmailQueueSize.Set(float64(s.queue.Len()))

			ctx := context.Background()
			start := time.Now()

			if err := s.emailService.SendEmail(ctx, job.Request); err != nil {
				s.log.Error("Failed to send bulk email", "error", err, "job_id", job.ID, "worker_id", id)
				metrics.FailedNotifications.WithLabelValues("email", job.Request.TenantID, "send_error").Inc()
			} else {
				metrics.NotificationsSent.WithLabelValues("email", job.Request.TenantID, "success").Inc()
			}

			// Record duration
			duration := time.Since(start).Seconds()
			metrics.NotificationDuration.WithLabelValues("email").Observe(duration)
		}
	}
}

// SendBulk queues bulk emails for sending
func (s *BulkEmailService) SendBulk(ctx context.Context, req *domain.BulkEmailRequest) error {
	s.log.Info("Queuing bulk emails", "tenant_id", req.TenantID, "recipients", len(req.Recipients))

	// Map request priority to queue priority
	var priority queue.Priority
	switch req.Priority {
	case 0:
		priority = queue.PriorityHigh
	case 1:
		priority = queue.PriorityNormal
	case 2:
		priority = queue.PriorityLow
	default:
		priority = queue.PriorityNormal
	}

	// Queue individual emails
	for _, recipient := range req.Recipients {
		emailReq := &domain.SendEmailRequest{
			TenantID:   req.TenantID,
			To:         []string{recipient},
			Subject:    req.Subject,
			Body:       req.Body,
			IsHTML:     req.IsHTML,
			TemplateID: req.TemplateID,
			Variables:  req.Variables,
		}

		job := &queue.EmailJob{
			ID:       uuid.New().String(),
			Priority: priority,
			Request:  emailReq,
		}

		s.queue.Push(job)
	}

	// Update queue size metric
	metrics.EmailQueueSize.Set(float64(s.queue.Len()))

	s.log.Info("Bulk emails queued", "count", len(req.Recipients), "queue_size", s.queue.Len())
	return nil
}

// QueueSize returns the current queue size
func (s *BulkEmailService) QueueSize() int {
	return s.queue.Len()
}
