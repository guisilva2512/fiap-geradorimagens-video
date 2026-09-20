package application

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	processingdomain "github.com/guisilva2512/fiap-geradorimagens-video/internal/video_processing/domain"
)

type workerRepositoryFake struct {
	processing *processingdomain.VideoProcessing
	statuses   []string
}

func (f *workerRepositoryFake) GetVideoProcessingByID(string, string) (*processingdomain.VideoProcessing, error) {
	return f.processing, nil
}

func (f *workerRepositoryFake) UpdateVideoProcessing(processing *processingdomain.VideoProcessing) error {
	f.statuses = append(f.statuses, processing.Status)
	return nil
}

type workerStorageFake struct {
	uploadedKeys []string
}

func (f *workerStorageFake) Download(_ context.Context, _ string, _ string, destination string) error {
	return os.WriteFile(destination, []byte("video"), 0o600)
}

func (f *workerStorageFake) Upload(_ context.Context, _ string, key string, _ string, _ string) error {
	f.uploadedKeys = append(f.uploadedKeys, key)
	return nil
}

type workerProcessorFake struct{}

func (workerProcessorFake) Process(_ context.Context, _ string, outputDir string) ([]string, error) {
	frame := filepath.Join(outputDir, "frame_000001.jpg")
	if err := os.WriteFile(frame, []byte("frame"), 0o600); err != nil {
		return nil, err
	}
	return []string{frame}, nil
}

func TestProcessRejectsInvalidPayload(t *testing.T) {
	service := NewService(&workerRepositoryFake{}, &workerStorageFake{}, workerProcessorFake{}, "bucket", t.TempDir())

	err := service.Process(context.Background(), []byte("invalid json"))
	if !errors.Is(err, ErrPermanent) {
		t.Fatalf("expected permanent error, got %v", err)
	}
}

func TestProcessCompletesValidMessage(t *testing.T) {
	repository := &workerRepositoryFake{processing: &processingdomain.VideoProcessing{
		ID: "processing-1", BatchID: "batch-1", Name: "video.mp4",
		StoragePath: "batch-1/processing-1/video.mp4", OutputPath: "batch-1/processing-1/output/",
	}}
	storage := &workerStorageFake{}
	service := NewService(repository, storage, workerProcessorFake{}, "bucket", t.TempDir())
	body, err := json.Marshal(processingdomain.VideoProcessing{
		ID: "processing-1", BatchID: "batch-1", StoragePath: "batch-1/processing-1/video.mp4",
	})
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	if err := service.Process(context.Background(), body); err != nil {
		t.Fatalf("expected processing to succeed: %v", err)
	}
	if len(repository.statuses) != 2 || repository.statuses[0] != "PROCESSING" || repository.statuses[1] != "COMPLETED" {
		t.Fatalf("unexpected status transitions: %v", repository.statuses)
	}
	if len(storage.uploadedKeys) != 1 || storage.uploadedKeys[0] != "batch-1/processing-1/output/frame_000001.jpg" {
		t.Fatalf("unexpected uploaded keys: %v", storage.uploadedKeys)
	}
}
