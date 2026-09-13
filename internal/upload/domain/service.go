package domain

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Definimos uma interface local ou usamos diretamente a do ports na injeção do main.
// Para blindar o service de importar o pacote ports, passamos a interface de repositório por parâmetro.
type UploadRepository interface {
	// VideoBatch methods
	List() ([]*VideoBatch, error)
	GetByID(id string) (*VideoBatch, error)
	Create(batch *VideoBatch) error
	Delete(id string) error

	// VideoProcessing methods
	ListVideoProcessings(batchId string) ([]*VideoProcessing, error)
	GetVideoProcessingByID(batchId string, id string) (*VideoProcessing, error)
	CreateVideoProcessing(processing *VideoProcessing) error
	UpdateVideoProcessing(processing *VideoProcessing) error
	DeleteVideoProcessing(batchId string, id string) error
}

type UploadStorage interface {
	SaveFile(bucket string, key string, file multipart.File) error
	GetFile(bucket string, key string) (io.ReadCloser, *int64, string, error)
}

type videoBatchService struct {
	repo    UploadRepository
	storage UploadStorage
}

// NewVideoBatchService retorna a struct concreta. No main.go, o Go vai aceitar
// essa struct como um ports.VideoBatchUseCase porque ela possui os métodos necessários.
func NewVideoBatchService(repo UploadRepository, storage UploadStorage) *videoBatchService {
	return &videoBatchService{repo: repo, storage: storage}
}

// VideoBatch methods
func (s *videoBatchService) List() ([]*VideoBatch, error) {
	return s.repo.List()
}

func (s *videoBatchService) Get(id string) (*VideoBatch, error) {
	existingBatch, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existingBatch == nil {
		return nil, errors.New("lote de vídeo não encontrado")
	}

	return existingBatch, nil
}

func (s *videoBatchService) Create(cmd CreateVideoBatchCommand) (*VideoBatch, error) {
	batch := &VideoBatch{
		ID:        uuid.New().String(),
		UserID:    cmd.UserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.repo.Create(batch)
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (s *videoBatchService) Delete(id string) error {
	existingBatch, err := s.repo.GetByID(id)

	if err != nil {
		return err
	}

	if existingBatch == nil {
		return errors.New("lote de vídeo não encontrado")
	}

	return s.repo.Delete(id)
}

// Processing methods
func (s *videoBatchService) ListProcessings(batchId string) ([]*VideoProcessing, error) {
	return s.repo.ListVideoProcessings(batchId)
}

func (s *videoBatchService) CreateProcessing(cmd CreateVideoProcessingCommand) (*VideoProcessing, error) {
	// Params for storage paths
	id := uuid.New().String()
	key := cmd.File.Filename
	storagePath := os.Getenv("AWS_S3_BUCKET") + "/" + cmd.BatchID + "/" + id + "/"
	outputPath := storagePath + "output/"

	if !isValidVideoFile(key) {
		return nil, errors.New("invalid video file format")
	}

	// Open the file
	file, err := cmd.File.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Save the file to storage
	err = s.storage.SaveFile(storagePath, key, file)
	if err != nil {
		return nil, err
	}

	// Create the VideoProcessing object
	processing := &VideoProcessing{
		ID:          id,
		BatchID:     cmd.BatchID,
		Status:      "PENDING",
		Name:        key,
		StoragePath: storagePath,
		OutputPath:  outputPath,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save the VideoProcessing object to the repository
	err = s.repo.CreateVideoProcessing(processing)
	if err != nil {
		return nil, err
	}

	return processing, nil
}

func isValidVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

func (s *videoBatchService) UpdateProcessing(cmd UpdateVideoProcessingCommand) (*VideoProcessing, error) {
	existingProcessing, err := s.repo.GetVideoProcessingByID(cmd.BatchID, cmd.ID)
	if err != nil {
		return nil, err
	}

	if existingProcessing == nil {
		return nil, errors.New("processamento de vídeo não encontrado")
	}

	// update the processing status and other fields as needed
	existingProcessing.Status = cmd.Status
	existingProcessing.UpdatedAt = time.Now()

	err = s.repo.UpdateVideoProcessing(existingProcessing)
	if err != nil {
		return nil, err
	}

	return existingProcessing, nil
}

func (s *videoBatchService) DownloadProcessing(batchId string, id string) (io.ReadCloser, *int64, string, string, error) {
	existingProcessing, err := s.repo.GetVideoProcessingByID(batchId, id)
	if err != nil {
		return nil, nil, "", "", err
	}
	if existingProcessing == nil {
		return nil, nil, "", "", errors.New("processamento de vídeo não encontrado")
	}

	output, contentLength, contentType, err := s.storage.GetFile(existingProcessing.StoragePath, existingProcessing.Name)
	if err != nil {
		return nil, nil, "", "", err
	}

	return output, contentLength, contentType, existingProcessing.Name, nil
}

func (s *videoBatchService) DeleteProcessing(batchId string, id string) error {
	existingProcessing, err := s.repo.GetVideoProcessingByID(batchId, id)
	if err != nil {
		return err
	}
	if existingProcessing == nil {
		return errors.New("processamento de vídeo não encontrado")
	}

	return s.repo.DeleteVideoProcessing(batchId, id)
}
