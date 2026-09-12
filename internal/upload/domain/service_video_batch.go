package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Definimos uma interface local ou usamos diretamente a do ports na injeção do main.
// Para blindar o service de importar o pacote ports, passamos a interface de repositório por parâmetro.
type internalRepository interface {
	List() ([]*VideoBatch, error)
	GetByID(id string) (*VideoBatch, error)
	Create(batch *VideoBatch) error
	Delete(id string) error
}

type videoBatchService struct {
	repo internalRepository
}

// NewVideoBatchService retorna a struct concreta. No main.go, o Go vai aceitar
// essa struct como um ports.VideoBatchUseCase porque ela possui os métodos necessários.
func NewVideoBatchService(repo internalRepository) *videoBatchService {
	return &videoBatchService{repo: repo}
}

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
