package messaging

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/guisilva2512/fiap-geradorimagens-video/internal/video_processing/application"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	connection *amqp.Connection
	queueName  string
	handler    *application.Service
}

func NewRabbitMQConsumer(connection *amqp.Connection, queueName string, handler *application.Service) *RabbitMQConsumer {
	return &RabbitMQConsumer{connection: connection, queueName: queueName, handler: handler}
}

func (c *RabbitMQConsumer) Run(ctx context.Context) error {
	channel, err := c.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if _, err := channel.QueueDeclare(c.queueName, true, false, false, false, nil); err != nil {
		return err
	}
	if err := channel.Qos(1, 0, false); err != nil {
		return err
	}

	messages, err := channel.Consume(c.queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message, ok := <-messages:
			if !ok {
				return errors.New("canal RabbitMQ encerrado")
			}

			err := c.handler.Process(ctx, message.Body)
			if err == nil {
				if err := message.Ack(false); err != nil {
					return fmt.Errorf("confirmar mensagem: %w", err)
				}
				continue
			}

			if errors.Is(err, application.ErrPermanent) {
				log.Printf("mensagem descartada: %v", err)
				if ackErr := message.Ack(false); ackErr != nil {
					return fmt.Errorf("confirmar mensagem inválida: %w", ackErr)
				}
				continue
			}

			log.Printf("falha transitória, reencaminhando mensagem: %v", err)
			if nackErr := message.Nack(false, true); nackErr != nil {
				return fmt.Errorf("reencaminhar mensagem: %w", nackErr)
			}
		}
	}
}
