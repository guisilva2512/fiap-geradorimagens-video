package messaging

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/guisilva2512/fiap-geradorimagens-video/internal/pkg/observability"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

type RabbitMQPublisher struct {
	mu         sync.Mutex
	connection *amqp.Connection
	url        string
	queueName  string
}

func NewRabbitMQPublisher(connection *amqp.Connection, queueName string) *RabbitMQPublisher {
	return &RabbitMQPublisher{connection: connection, queueName: queueName}
}

func NewRabbitMQPublisherWithReconnect(connection *amqp.Connection, url string, queueName string) *RabbitMQPublisher {
	return &RabbitMQPublisher{connection: connection, url: url, queueName: queueName}
}

func (p *RabbitMQPublisher) CreateQueue(ctx context.Context) error {
	connection, err := p.ensureConnection()
	if err != nil {
		return err
	}
	channel, err := connection.Channel()
	if err != nil {
		p.invalidateConnection(connection)
		return err
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(p.queueName, true, false, false, false, nil)
	return err
}

func (p *RabbitMQPublisher) PublishMessage(ctx context.Context, body []byte) error {
	ctx, span := otel.Tracer("upload-api").Start(ctx, "rabbitmq.publish")
	defer span.End()

	var lastErr error
	for attempt := 1; attempt <= 2; attempt++ {
		lastErr = p.publishOnce(ctx, body)
		if lastErr == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return fmt.Errorf("publicar mensagem no RabbitMQ após reconexão: %w", lastErr)
}

func (p *RabbitMQPublisher) publishOnce(ctx context.Context, body []byte) error {
	connection, err := p.ensureConnection()
	if err != nil {
		return err
	}

	channel, err := connection.Channel()
	if err != nil {
		p.invalidateConnection(connection)
		return err
	}
	defer channel.Close()

	if err := channel.Confirm(false); err != nil {
		p.invalidateConnection(connection)
		return err
	}

	confirmations := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	returns := channel.NotifyReturn(make(chan amqp.Return, 1))

	headers := amqp.Table{}
	observability.InjectAMQP(ctx, &headers)
	if err := channel.PublishWithContext(ctx, "", p.queueName, true, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
		Body:         body,
	}); err != nil {
		p.invalidateConnection(connection)
		return err
	}

	select {
	case returned := <-returns:
		return fmt.Errorf("mensagem não roteada pelo RabbitMQ: %s", returned.ReplyText)
	case confirmation := <-confirmations:
		if !confirmation.Ack {
			return errors.New("RabbitMQ rejeitou a mensagem")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *RabbitMQPublisher) ensureConnection() (*amqp.Connection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.connection != nil && !p.connection.IsClosed() {
		return p.connection, nil
	}
	if p.url == "" {
		return nil, errors.New("conexão RabbitMQ encerrada e URL de reconexão não configurada")
	}

	connection, err := amqp.Dial(p.url)
	if err != nil {
		return nil, err
	}
	p.connection = connection
	return connection, nil
}

func (p *RabbitMQPublisher) invalidateConnection(connection *amqp.Connection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.connection == connection {
		p.connection = nil
	}
}
