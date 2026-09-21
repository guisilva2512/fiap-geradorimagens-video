package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	database "github.com/guisilva2512/fiap-geradorimagens-video/internal/pkg/databases"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/pkg/observability"
	workerstorage "github.com/guisilva2512/fiap-geradorimagens-video/internal/pkg/storage"
	uploaddb "github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/adapters/db"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/video_processing/adapters/messaging"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/video_processing/adapters/processor"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/video_processing/application"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	shutdownTracer, err := observability.InitTracer(context.Background(), "video-worker")
	if err != nil {
		log.Fatalf("inicializar OpenTelemetry: %v", err)
	}
	defer shutdownTracer(context.Background())

	queueName := requiredEnv("RABBITMQ_QUEUE")
	connection, err := amqp.Dial(requiredEnv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("conectar ao RabbitMQ: %v", err)
	}
	defer connection.Close()

	workerStorage := newStorage()
	repository := uploaddb.NewPostgresVideoBatchRepository(database.Databases())
	service := application.NewService(repository, workerStorage, processor.NewFFmpeg(getenv("FFMPEG_BINARY", "ffmpeg"), getenv("VIDEO_FRAMES_PER_SECOND", "1")), requiredEnv("AWS_S3_BUCKET"), getenv("WORKER_TEMP_DIR", os.TempDir()))
	consumer := messaging.NewRabbitMQConsumer(connection, queueName, service)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metricsServer := &http.Server{Addr: ":8083", Handler: observability.MetricsHandler()}
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("servidor de metricas encerrado: %v", err)
		}
	}()
	defer metricsServer.Shutdown(context.Background())

	log.Printf("video worker consumindo a fila %q", queueName)
	if err := consumer.Run(ctx); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
}

func newStorage() *workerstorage.S3 {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(requiredEnv("AWS_REGION")), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(requiredEnv("AWS_ACCESS_KEY_ID"), requiredEnv("AWS_SECRET_ACCESS_KEY"), "")))
	if err != nil {
		log.Fatalf("carregar configuração AWS: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(requiredEnv("AWS_S3_ENDPOINT"))
		options.UsePathStyle = true
	})
	return workerstorage.NewS3(client)
}

func requiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s não configurada", name)
	}
	return value
}

func getenv(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
