package messaging

import (
	"fmt"
	"snowflake/pkg/logger"

	"go.uber.org/zap"

	"github.com/gofiber/fiber/v2/log"
	"github.com/rabbitmq/amqp091-go"
)

// MessagingService holds the RabbitMQ connection and logger
type MessagingService struct {
	conn   *amqp091.Connection
	logger *logger.Logger
}

// NewRabbitMQ initializes the connection and returns the MessagingService
func NewRabbitMQ(url string, logger *logger.Logger) (*MessagingService, error) {
	// Establish connection to RabbitMQ
	conn, err := amqp091.Dial(url)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	logger.Info("Connected to RabbitMQ")

	// Return the initialized service
	return &MessagingService{
		conn:   conn,
		logger: logger,
	}, nil
}

// Produce sends a message to a specific queue
func (m *MessagingService) Produce(queueName, message string) error {
	// Create a new channel for the operation
	channel, err := m.conn.Channel()
	if err != nil {
		m.logger.Error("Failed to open a channel", zap.Error(err))
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer channel.Close()

	// Declare the queue
	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		m.logger.Error("Failed to declare queue", zap.Error(err), zap.String("queueName", queueName))
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	// Publish the message
	err = channel.Publish(
		"",         // exchange
		queue.Name, // routing key (queue name)
		true,       // mandatory, ensures message will not be dropped
		false,      // immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		m.logger.Error("Failed to send message", zap.Error(err), zap.String("queueName", queueName), zap.String("message", message))
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Info("Sent message", zap.String("queueName", queueName), zap.String("message", message))

	return nil
}

// Consume listens for messages from a specific queue
func (m *MessagingService) Consume(queueName string) (<-chan amqp091.Delivery, error) {
	// Create a new channel for the operation
	channel, err := m.conn.Channel()
	if err != nil {
		m.logger.Error("Failed to open a channel", zap.Error(err))
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare the queue
	_, err = channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		m.logger.Error("Failed to declare queue", zap.Error(err), zap.String("queueName", queueName))
		channel.Close()
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	// Start consuming messages
	messages, err := channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		m.logger.Error("Failed to start consuming messages", zap.Error(err), zap.String("queueName", queueName))
		channel.Close()
		return nil, fmt.Errorf("failed to start consuming messages: %w", err)
	}

	m.logger.Info("Started consuming messages", zap.String("queueName", queueName))
	return messages, nil
}

// Close gracefully closes the connection
func (m *MessagingService) Close() {
	m.logger.Info("Closing RabbitMQ connection")
	m.conn.Close()
}
