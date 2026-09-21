package messaging

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/guisilva2512/fiap-geradorimagens-video/internal/pkg/observability"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/video_processing/application"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
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

			log.Printf("rabbitmq_message_received queue=%s delivery_tag=%d body_bytes=%d", c.queueName, message.DeliveryTag, len(message.Body))
			messageContext := observability.ExtractAMQP(ctx, message.Headers)
			messageContext, span := otel.Tracer("video-worker").Start(messageContext, "rabbitmq.consume")
			started := time.Now()
			err := c.handler.Process(messageContext, message.Body)
			span.End()
			if err == nil {
				observability.RecordWorkerMessage("video-worker", "completed", time.Since(started))
				log.Printf("rabbitmq_message_processed queue=%s delivery_tag=%d duration=%s action=ack", c.queueName, message.DeliveryTag, time.Since(started))
				if err := message.Ack(false); err != nil {
					log.Printf("rabbitmq_ack_failed queue=%s delivery_tag=%d error=%q", c.queueName, message.DeliveryTag, err)
					return fmt.Errorf("confirmar mensagem: %w", err)
				}
				continue
			}

			if errors.Is(err, application.ErrPermanent) {
				observability.RecordWorkerMessage("video-worker", "failed", time.Since(started))
				log.Printf("rabbitmq_message_discarded queue=%s delivery_tag=%d duration=%s action=ack error=%q", c.queueName, message.DeliveryTag, time.Since(started), err)
				if ackErr := message.Ack(false); ackErr != nil {
					log.Printf("rabbitmq_ack_failed queue=%s delivery_tag=%d error=%q", c.queueName, message.DeliveryTag, ackErr)
					return fmt.Errorf("confirmar mensagem inválida: %w", ackErr)
				}
				continue
			}

			observability.RecordWorkerMessage("video-worker", "retried", time.Since(started))
			log.Printf("rabbitmq_message_requeued queue=%s delivery_tag=%d duration=%s action=nack_requeue error=%q", c.queueName, message.DeliveryTag, time.Since(started), err)
			if nackErr := message.Nack(false, true); nackErr != nil {
				log.Printf("rabbitmq_nack_failed queue=%s delivery_tag=%d error=%q", c.queueName, message.DeliveryTag, nackErr)
				return fmt.Errorf("reencaminhar mensagem: %w", nackErr)
			}
		}
	}
}
