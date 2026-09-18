package domain

import (
	"mime/multipart"

	processingdomain "github.com/guisilva2512/fiap-geradorimagens-video/internal/video_processing/domain"
)

type VideoProcessing = processingdomain.VideoProcessing

type CreateVideoProcessingCommand struct {
	BatchID string
	File    *multipart.FileHeader
	// Status      string
	// Name        string
	// StoragePath string
	// OutputPath  string

}

type UpdateVideoProcessingCommand struct {
	ID      string
	BatchID string
	Status  string
}
