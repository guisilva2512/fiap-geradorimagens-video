package ports

import "github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/domain"

type VideoBatchUseCase interface {
	List() ([]*domain.VideoBatch, error)
	Get(id string) (*domain.VideoBatch, error)
	Create(cmd domain.CreateVideoBatchCommand) (*domain.VideoBatch, error)
	Delete(id string) error
}
