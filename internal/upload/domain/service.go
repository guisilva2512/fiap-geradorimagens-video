package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Definimos uma interface local ou usamos diretamente a do ports na injeção do main.
// Para blindar o service de importar o pacote ports, passamos a interface de repositório por parâmetro.
type internalRepository interface {
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

type videoBatchService struct {
	repo internalRepository
}

// NewVideoBatchService retorna a struct concreta. No main.go, o Go vai aceitar
// essa struct como um ports.VideoBatchUseCase porque ela possui os métodos necessários.
func NewVideoBatchService(repo internalRepository) *videoBatchService {
	return &videoBatchService{repo: repo}
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

	return s.repo.GetByID(id)
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
	processing := &VideoProcessing{
		ID:          uuid.New().String(),
		BatchID:     cmd.BatchID,
		Status:      cmd.Status,
		Name:        cmd.Name,
		StoragePath: cmd.StoragePath,
		OutputPath:  cmd.OutputPath,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.repo.CreateVideoProcessing(processing)
	if err != nil {
		return nil, err
	}

	return processing, nil
}

func (s *videoBatchService) UpdateProcessing(cmd UpdateVideoProcessingCommand) (*VideoProcessing, error) {
	existingProcessing, err := s.repo.GetVideoProcessingByID(cmd.BatchID, cmd.ID)
	if err != nil {
		return nil, err
	}

	if existingProcessing == nil {
		return nil, errors.New("processamento de vídeo não encontrado")
	}

	existingProcessing.Status = cmd.Status
	existingProcessing.UpdatedAt = time.Now()

	err = s.repo.UpdateVideoProcessing(existingProcessing)
	if err != nil {
		return nil, err
	}

	return existingProcessing, nil
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
