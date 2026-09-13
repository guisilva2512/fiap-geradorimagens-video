package domain

import (
	"mime/multipart"
	"time"
)

type VideoProcessing struct {
	ID           string
	BatchID      string
	Status       string
	Name         string
	StoragePath  string
	OutputPath   string
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

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
