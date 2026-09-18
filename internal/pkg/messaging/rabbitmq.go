package messaging

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	connection *amqp.Connection
	queueName  string
}

func NewRabbitMQPublisher(connection *amqp.Connection, queueName string) *RabbitMQPublisher {
	return &RabbitMQPublisher{connection: connection, queueName: queueName}
}

func (p *RabbitMQPublisher) CreateQueue(ctx context.Context) error {
	channel, err := p.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(p.queueName, true, false, false, false, nil)
	return err
}

func (p *RabbitMQPublisher) PublishMessage(ctx context.Context, body []byte) error {
	channel, err := p.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := channel.Confirm(false); err != nil {
		return err
	}

	confirmations := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	returns := channel.NotifyReturn(make(chan amqp.Return, 1))

	if err := channel.PublishWithContext(ctx, "", p.queueName, true, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	}); err != nil {
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
