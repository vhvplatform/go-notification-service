package consumer

import (
	"context"
	"encoding/json"

	"github.com/longvhv/saas-framework-go/pkg/logger"
	"github.com/longvhv/saas-framework-go/pkg/rabbitmq"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/domain"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/service"
)

const (
	notificationExchange = "notifications"
	notificationQueue    = "notification_queue"
	notificationRoutingKey = "notification.*"
)

// EventConsumer consumes events from RabbitMQ
type EventConsumer struct {
	client  *rabbitmq.RabbitMQClient
	service *service.NotificationService
	log     *logger.Logger
}

// NewEventConsumer creates a new event consumer
func NewEventConsumer(client *rabbitmq.RabbitMQClient, service *service.NotificationService, log *logger.Logger) *EventConsumer {
	return &EventConsumer{
		client:  client,
		service: service,
		log:     log,
	}
}

// Start starts consuming events from RabbitMQ
func (c *EventConsumer) Start() error {
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
	messages, err := c.client.Consume(notificationQueue)
	if err != nil {
		c.log.Error("Failed to start consuming", "error", err)
		return err
	}

	// Process messages
	for msg := range messages {
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

	return nil
}
