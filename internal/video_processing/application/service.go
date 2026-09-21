package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	processingdomain "github.com/guisilva2512/fiap-geradorimagens-video/internal/video_processing/domain"
	"go.opentelemetry.io/otel/trace"
)

var ErrPermanent = errors.New("erro permanente no processamento")

type Repository interface {
	GetVideoProcessingByID(batchID string, id string) (*processingdomain.VideoProcessing, error)
	UpdateVideoProcessing(processing *processingdomain.VideoProcessing) error
}

type Storage interface {
	Download(ctx context.Context, bucket string, key string, destination string) error
	Upload(ctx context.Context, bucket string, key string, source string, contentType string) error
}

type Processor interface {
	Process(ctx context.Context, input string, outputDir string) ([]string, error)
}

type Service struct {
	repository Repository
	storage    Storage
	processor  Processor
	bucket     string
	tempDir    string
}

func NewService(repository Repository, storage Storage, processor Processor, bucket string, tempDir string) *Service {
	return &Service{
		repository: repository,
		storage:    storage,
		processor:  processor,
		bucket:     bucket,
		tempDir:    tempDir,
	}
}

func (s *Service) Process(ctx context.Context, body []byte) error {
	started := time.Now()
	logContext(ctx, "processing_started payload_bytes=%d", len(body))

	var message processingdomain.VideoProcessing
	if err := json.Unmarshal(body, &message); err != nil {
		logContext(ctx, "processing_rejected reason=invalid_payload error=%q", err)
		return fmt.Errorf("%w: payload RabbitMQ inválido: %v", ErrPermanent, err)
	}
	if message.ID == "" || message.BatchID == "" || message.StoragePath == "" {
		logContext(ctx, "processing_rejected reason=missing_required_fields processing_id=%q batch_id=%q storage_path=%q", message.ID, message.BatchID, message.StoragePath)
		return fmt.Errorf("%w: payload sem id, batch_id ou storage_path", ErrPermanent)
	}
	logContext(ctx, "processing_message_received processing_id=%s batch_id=%s storage_path=%s output_path=%s", message.ID, message.BatchID, message.StoragePath, message.OutputPath)

	stepStarted := time.Now()
	processing, err := s.repository.GetVideoProcessingByID(message.BatchID, message.ID)
	if err != nil {
		logContext(ctx, "processing_repository_lookup_failed processing_id=%s batch_id=%s error=%q", message.ID, message.BatchID, err)
		return fmt.Errorf("buscar processamento: %w", err)
	}
	if processing == nil {
		logContext(ctx, "processing_rejected reason=not_found processing_id=%s batch_id=%s", message.ID, message.BatchID)
		return fmt.Errorf("%w: processamento %s não encontrado", ErrPermanent, message.ID)
	}
	logContext(ctx, "processing_repository_lookup_completed processing_id=%s duration=%s current_status=%s", processing.ID, time.Since(stepStarted), processing.Status)

	processing.Status = "PROCESSING"
	processing.ErrorMessage = ""
	processing.UpdatedAt = time.Now()
	if err := s.repository.UpdateVideoProcessing(processing); err != nil {
		logContext(ctx, "processing_status_update_failed processing_id=%s status=PROCESSING error=%q", processing.ID, err)
		return fmt.Errorf("marcar processamento como PROCESSING: %w", err)
	}
	logContext(ctx, "processing_status_updated processing_id=%s status=PROCESSING", processing.ID)

	stepStarted = time.Now()
	if err := s.process(ctx, processing); err != nil {
		logContext(ctx, "processing_pipeline_failed processing_id=%s duration=%s error=%q", processing.ID, time.Since(stepStarted), err)
		processing.Status = "FAILED"
		processing.ErrorMessage = err.Error()
		processing.UpdatedAt = time.Now()
		if updateErr := s.repository.UpdateVideoProcessing(processing); updateErr != nil {
			logContext(ctx, "processing_status_update_failed processing_id=%s status=FAILED error=%q update_error=%q", processing.ID, err, updateErr)
			return fmt.Errorf("%w; atualizar status FAILED: %v", err, updateErr)
		}
		logContext(ctx, "processing_status_updated processing_id=%s status=FAILED error=%q", processing.ID, err)
		return fmt.Errorf("%w: %v", ErrPermanent, err)
	}
	logContext(ctx, "processing_pipeline_completed processing_id=%s duration=%s", processing.ID, time.Since(stepStarted))

	processing.Status = "COMPLETED"
	processing.ErrorMessage = ""
	processing.UpdatedAt = time.Now()
	if err := s.repository.UpdateVideoProcessing(processing); err != nil {
		logContext(ctx, "processing_status_update_failed processing_id=%s status=COMPLETED error=%q", processing.ID, err)
		return fmt.Errorf("marcar processamento como COMPLETED: %w", err)
	}

	logContext(ctx, "processing_completed processing_id=%s batch_id=%s total_duration=%s status=COMPLETED", processing.ID, processing.BatchID, time.Since(started))

	return nil
}

func (s *Service) process(ctx context.Context, processing *processingdomain.VideoProcessing) error {
	started := time.Now()
	logContext(ctx, "pipeline_started processing_id=%s batch_id=%s temp_dir=%s", processing.ID, processing.BatchID, s.tempDir)

	if err := os.MkdirAll(s.tempDir, 0o755); err != nil {
		logContext(ctx, "pipeline_failed step=create_temp_dir processing_id=%s error=%q", processing.ID, err)
		return fmt.Errorf("criar diretório temporário: %w", err)
	}

	workDir, err := os.MkdirTemp(s.tempDir, "video-processing-")
	if err != nil {
		logContext(ctx, "pipeline_failed step=create_workspace processing_id=%s error=%q", processing.ID, err)
		return fmt.Errorf("criar workspace temporário: %w", err)
	}
	defer os.RemoveAll(workDir)
	logContext(ctx, "workspace_created processing_id=%s path=%s", processing.ID, workDir)

	inputPath := filepath.Join(workDir, processing.Name)
	outputDir := filepath.Join(workDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		logContext(ctx, "pipeline_failed step=create_output_dir processing_id=%s error=%q", processing.ID, err)
		return fmt.Errorf("criar diretório de imagens: %w", err)
	}

	stepStarted := time.Now()
	logContext(ctx, "video_download_started processing_id=%s bucket=%s key=%s destination=%s", processing.ID, s.bucket, processing.StoragePath, inputPath)
	if err := s.storage.Download(ctx, s.bucket, processing.StoragePath, inputPath); err != nil {
		logContext(ctx, "video_download_failed processing_id=%s duration=%s error=%q", processing.ID, time.Since(stepStarted), err)
		return fmt.Errorf("baixar vídeo original: %w", err)
	}
	logContext(ctx, "video_download_completed processing_id=%s duration=%s", processing.ID, time.Since(stepStarted))

	stepStarted = time.Now()
	logContext(ctx, "ffmpeg_started processing_id=%s input=%s output_dir=%s", processing.ID, inputPath, outputDir)
	files, err := s.processor.Process(ctx, inputPath, outputDir)
	if err != nil {
		logContext(ctx, "ffmpeg_failed processing_id=%s duration=%s error=%q", processing.ID, time.Since(stepStarted), err)
		return fmt.Errorf("processar vídeo: %w", err)
	}
	logContext(ctx, "ffmpeg_completed processing_id=%s duration=%s images=%d", processing.ID, time.Since(stepStarted), len(files))

	for _, file := range files {
		outputKey := filepath.ToSlash(filepath.Join(processing.OutputPath, filepath.Base(file)))
		stepStarted = time.Now()
		logContext(ctx, "image_upload_started processing_id=%s file=%s key=%s", processing.ID, file, outputKey)
		if err := s.storage.Upload(ctx, s.bucket, outputKey, file, "image/jpeg"); err != nil {
			logContext(ctx, "image_upload_failed processing_id=%s key=%s duration=%s error=%q", processing.ID, outputKey, time.Since(stepStarted), err)
			return fmt.Errorf("enviar imagem processada: %w", err)
		}
		logContext(ctx, "image_upload_completed processing_id=%s key=%s duration=%s", processing.ID, outputKey, time.Since(stepStarted))
	}
	logContext(ctx, "pipeline_completed processing_id=%s images=%d duration=%s", processing.ID, len(files), time.Since(started))
	return nil
}

func logContext(ctx context.Context, format string, args ...interface{}) {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		log.Printf("trace_id=%s span_id=%s %s", spanContext.TraceID(), spanContext.SpanID(), fmt.Sprintf(format, args...))
		return
	}
	log.Printf("%s", fmt.Sprintf(format, args...))
}
