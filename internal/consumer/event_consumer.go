package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/service"
	"github.com/vhvplatform/go-notification-service/internal/shared/logger"
	"github.com/vhvplatform/go-notification-service/internal/shared/rabbitmq"
)

const (
	notificationExchange   = "notifications"
	notificationQueue      = "notification_queue"
	notificationRoutingKey = "notification.*"
)

// EventConsumer consumes events from RabbitMQ
type EventConsumer struct {
	client        *rabbitmq.RabbitMQClient
	service       *service.NotificationService
	log           *logger.Logger
	stopChan      chan struct{}
	maxRetries    int
	retryDelay    time.Duration
	maxRetryDelay time.Duration
}

// NewEventConsumer creates a new event consumer
func NewEventConsumer(client *rabbitmq.RabbitMQClient, service *service.NotificationService, log *logger.Logger) *EventConsumer {
	return &EventConsumer{
		client:        client,
		service:       service,
		log:           log,
		stopChan:      make(chan struct{}),
		maxRetries:    5,
		retryDelay:    1 * time.Second,
		maxRetryDelay: 60 * time.Second,
	}
}

// Start starts consuming events from RabbitMQ with auto-restart
func (c *EventConsumer) Start() error {
	c.log.Info("Starting event consumer with auto-restart", "queue", notificationQueue)

	// Run consumer with exponential backoff retry
	go c.runWithRetry()

	return nil
}

// Stop stops the event consumer
func (c *EventConsumer) Stop() {
	close(c.stopChan)
}

// runWithRetry runs the consumer with exponential backoff retry
func (c *EventConsumer) runWithRetry() {
	retryCount := 0
	currentDelay := c.retryDelay

	for {
		select {
		case <-c.stopChan:
			c.log.Info("Event consumer stopped")
			return
		default:
			err := c.consume()
			if err != nil {
				retryCount++
				c.log.Error("Consumer failed, retrying", "error", err, "retry_count", retryCount, "delay", currentDelay)

				// Wait before retry
				time.Sleep(currentDelay)

				// Calculate next delay with exponential backoff
				currentDelay = currentDelay * 2
				if currentDelay > c.maxRetryDelay {
					currentDelay = c.maxRetryDelay
				}
			} else {
				// Reset retry count on successful run
				retryCount = 0
				currentDelay = c.retryDelay
			}
		}
	}
}

// consume performs the actual consumption of messages
func (c *EventConsumer) consume() error {
	c.log.Info("Starting event consumer", "queue", notificationQueue)

	// Declare exchange
	if err := c.client.DeclareExchange(notificationExchange, "topic"); err != nil {
		c.log.Error("Failed to declare exchange", "error", err)
		return err
	}

	// Declare queue
	if err := c.client.DeclareQueue(notificationQueue); err != nil {
		c.log.Error("Failed to declare queue", "error", err)
		return err
	}

	// Bind queue to exchange
	if err := c.client.BindQueue(notificationQueue, notificationRoutingKey, notificationExchange); err != nil {
		c.log.Error("Failed to bind queue", "error", err)
		return err
	}

	// Start consuming
	messages, err := c.client.Consume(notificationQueue, notificationRoutingKey)
	if err != nil {
		c.log.Error("Failed to start consuming", "error", err)
		return err
	}

	// Process messages
	for msg := range messages {
		select {
		case <-c.stopChan:
			return nil
		default:
			c.log.Info("Received message", "routing_key", msg.RoutingKey)

			var event domain.Event
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				c.log.Error("Failed to unmarshal event", "error", err)
				msg.Nack(false, false) // Don't requeue invalid messages
				continue
			}

			// Process event
			ctx := context.Background()
			if err := c.service.ProcessEvent(ctx, &event); err != nil {
				c.log.Error("Failed to process event", "error", err, "type", event.Type)
				msg.Nack(false, true) // Requeue for retry
				continue
			}

			// Acknowledge message
			msg.Ack(false)
			c.log.Info("Event processed successfully", "type", event.Type)
		}
	}

	return nil
}
