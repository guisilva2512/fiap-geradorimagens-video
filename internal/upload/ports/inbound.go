package ports

import (
	"context"
	"io"

	"github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/domain"
)

type VideoBatchUseCase interface {
	// VideoBatch methods
	List() ([]*domain.VideoBatch, error)
	Get(id string) (*domain.VideoBatch, error)
	Create(cmd domain.CreateVideoBatchCommand) (*domain.VideoBatch, error)
	Delete(id string) error

	// Processing methods
	ListProcessings(batchId string) ([]*domain.VideoProcessing, error)
	CreateProcessing(ctx context.Context, cmd domain.CreateVideoProcessingCommand) (*domain.VideoProcessing, error)
	UpdateProcessing(cmd domain.UpdateVideoProcessingCommand) (*domain.VideoProcessing, error)
	DeleteProcessing(batchId string, id string) error
	DownloadProcessing(batchId string, id string) (io.ReadCloser, *int64, string, string, error)
	WriteImagesZip(ctx context.Context, batchID string, processingID string, writer io.Writer) error
}
