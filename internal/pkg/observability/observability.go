package observability

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fiap",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total de requisicoes HTTP processadas.",
		},
		[]string{"service", "method", "route", "status"},
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fiap",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Duracao das requisicoes HTTP em segundos.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service", "method", "route"},
	)
	workerMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fiap",
			Subsystem: "worker",
			Name:      "messages_total",
			Help:      "Total de mensagens RabbitMQ processadas pelo worker.",
		},
		[]string{"service", "status"},
	)
	workerDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fiap",
			Subsystem: "worker",
			Name:      "processing_duration_seconds",
			Help:      "Duracao do processamento de mensagens em segundos.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration, workerMessages, workerDuration)
}

func InitTracer(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector:4318"
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	resourceAttributes, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resourceAttributes),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return provider.Shutdown, nil
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func MetricsMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		httpRequests.WithLabelValues(serviceName, c.Request.Method, route, status).Inc()
		httpDuration.WithLabelValues(serviceName, c.Request.Method, route).Observe(time.Since(started).Seconds())
	}
}

func RecordWorkerMessage(serviceName, status string, duration time.Duration) {
	workerMessages.WithLabelValues(serviceName, status).Inc()
	workerDuration.WithLabelValues(serviceName).Observe(duration.Seconds())
}

type AMQPHeaderCarrier amqp091.Table

func (c AMQPHeaderCarrier) Get(key string) string {
	value, ok := c[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return ""
	}
}

func (c AMQPHeaderCarrier) Set(key string, value string) {
	c[key] = value
}

func (c AMQPHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for key := range c {
		keys = append(keys, key)
	}
	return keys
}

func InjectAMQP(ctx context.Context, headers *amqp091.Table) {
	if *headers == nil {
		*headers = amqp091.Table{}
	}
	otel.GetTextMapPropagator().Inject(ctx, AMQPHeaderCarrier(*headers))
}

func ExtractAMQP(ctx context.Context, headers amqp091.Table) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, AMQPHeaderCarrier(headers))
}
