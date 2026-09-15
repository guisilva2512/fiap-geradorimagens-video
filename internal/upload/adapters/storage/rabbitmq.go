package storage

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQStorageAdapter struct {
	connection *amqp.Connection
	queueName  string
}

func NewRabbitMQStorageAdapter(connection *amqp.Connection, queueName string) *RabbitMQStorageAdapter {
	return &RabbitMQStorageAdapter{connection: connection, queueName: queueName}
}

func (r *RabbitMQStorageAdapter) CreateQueue(ctx context.Context) error {
	ch, err := r.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		r.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQStorageAdapter) PublishMessage(
	ctx context.Context,
	body []byte,
) error {
	ch, err := r.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.Confirm(false); err != nil {
		return err
	}

	confirmations := ch.NotifyPublish(
		make(chan amqp.Confirmation, 1),
	)

	returns := ch.NotifyReturn(
		make(chan amqp.Return, 1),
	)

	err = ch.PublishWithContext(
		ctx,
		"",
		r.queueName,
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return err
	}

	select {
	case returned := <-returns:
		return fmt.Errorf(
			"mensagem não roteada pelo RabbitMQ: %s",
			returned.ReplyText,
		)

	case confirmation := <-confirmations:
		if !confirmation.Ack {
			return errors.New("RabbitMQ rejeitou a mensagem")
		}

		// Para mensagens roteadas, não haverá retorno.
		// Para mensagens não roteadas, o Basic.Return deve chegar antes
		// da confirmação de publicação.
		select {
		case returned := <-returns:
			return fmt.Errorf(
				"mensagem não roteada pelo RabbitMQ: %s",
				returned.ReplyText,
			)
		default:
			return nil
		}

	case <-ctx.Done():
		return ctx.Err()
	}
}
