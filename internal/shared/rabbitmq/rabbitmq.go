package rabbitmq

import (
	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQClient wraps the RabbitMQ connection
type RabbitMQClient struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

// Message represents a RabbitMQ message
type Message struct {
	Body       []byte
	RoutingKey string
	delivery   amqp091.Delivery
}

// Ack acknowledges a message
func (m *Message) Ack(multiple bool) error {
	return m.delivery.Ack(multiple)
}

// Nack negative acknowledges a message
func (m *Message) Nack(multiple, requeue bool) error {
	return m.delivery.Nack(multiple, requeue)
}

// NewRabbitMQClient creates a new RabbitMQ client
func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &RabbitMQClient{
		conn:    conn,
		channel: channel,
	}, nil
}

// DeclareExchange declares an exchange
func (c *RabbitMQClient) DeclareExchange(name, kind string) error {
	return c.channel.ExchangeDeclare(
		name,
		kind,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
}

// DeclareQueue declares a queue
func (c *RabbitMQClient) DeclareQueue(name string) error {
	_, err := c.channel.QueueDeclare(
		name,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}

// BindQueue binds a queue to an exchange
func (c *RabbitMQClient) BindQueue(queue, routingKey, exchange string) error {
	return c.channel.QueueBind(
		queue,
		routingKey,
		exchange,
		false, // no-wait
		nil,   // arguments
	)
}

// Consume starts consuming messages from a queue
func (c *RabbitMQClient) Consume(queue, consumerTag string) (<-chan Message, error) {
	msgs, err := c.channel.Consume(
		queue,
		consumerTag,
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	// Convert to our Message type
	messageChan := make(chan Message)
	go func() {
		for d := range msgs {
			messageChan <- Message{
				Body:       d.Body,
				RoutingKey: d.RoutingKey,
				delivery:   d,
			}
		}
		close(messageChan)
	}()

	return messageChan, nil
}

// Publish publishes a message to an exchange
func (c *RabbitMQClient) Publish(exchange, routingKey string, body []byte) error {
	return c.channel.Publish(
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// Close closes the RabbitMQ connection
func (c *RabbitMQClient) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
