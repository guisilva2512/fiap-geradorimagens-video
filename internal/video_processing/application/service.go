package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	processingdomain "github.com/guisilva2512/fiap-geradorimagens-video/internal/video_processing/domain"
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
	var message processingdomain.VideoProcessing
	if err := json.Unmarshal(body, &message); err != nil {
		return fmt.Errorf("%w: payload RabbitMQ inválido: %v", ErrPermanent, err)
	}
	if message.ID == "" || message.BatchID == "" || message.StoragePath == "" {
		return fmt.Errorf("%w: payload sem id, batch_id ou storage_path", ErrPermanent)
	}

	processing, err := s.repository.GetVideoProcessingByID(message.BatchID, message.ID)
	if err != nil {
		return fmt.Errorf("buscar processamento: %w", err)
	}
	if processing == nil {
		return fmt.Errorf("%w: processamento %s não encontrado", ErrPermanent, message.ID)
	}

	processing.Status = "PROCESSING"
	processing.ErrorMessage = ""
	processing.UpdatedAt = time.Now()
	if err := s.repository.UpdateVideoProcessing(processing); err != nil {
		return fmt.Errorf("marcar processamento como PROCESSING: %w", err)
	}

	if err := s.process(ctx, processing); err != nil {
		processing.Status = "FAILED"
		processing.ErrorMessage = err.Error()
		processing.UpdatedAt = time.Now()
		if updateErr := s.repository.UpdateVideoProcessing(processing); updateErr != nil {
			return fmt.Errorf("%w; atualizar status FAILED: %v", err, updateErr)
		}
		return fmt.Errorf("%w: %v", ErrPermanent, err)
	}

	processing.Status = "COMPLETED"
	processing.ErrorMessage = ""
	processing.UpdatedAt = time.Now()
	if err := s.repository.UpdateVideoProcessing(processing); err != nil {
		return fmt.Errorf("marcar processamento como COMPLETED: %w", err)
	}

	return nil
}

func (s *Service) process(ctx context.Context, processing *processingdomain.VideoProcessing) error {
	if err := os.MkdirAll(s.tempDir, 0o755); err != nil {
		return fmt.Errorf("criar diretório temporário: %w", err)
	}

	workDir, err := os.MkdirTemp(s.tempDir, "video-processing-")
	if err != nil {
		return fmt.Errorf("criar workspace temporário: %w", err)
	}
	defer os.RemoveAll(workDir)

	inputPath := filepath.Join(workDir, processing.Name)
	outputDir := filepath.Join(workDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("criar diretório de imagens: %w", err)
	}
	if err := s.storage.Download(ctx, s.bucket, processing.StoragePath, inputPath); err != nil {
		return fmt.Errorf("baixar vídeo original: %w", err)
	}
	files, err := s.processor.Process(ctx, inputPath, outputDir)
	if err != nil {
		return fmt.Errorf("processar vídeo: %w", err)
	}

	for _, file := range files {
		outputKey := filepath.ToSlash(filepath.Join(processing.OutputPath, filepath.Base(file)))
		if err := s.storage.Upload(ctx, s.bucket, outputKey, file, "image/jpeg"); err != nil {
			return fmt.Errorf("enviar imagem processada: %w", err)
		}
	}
	return nil
}
