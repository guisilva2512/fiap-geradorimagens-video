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

type failingRepositoryFake struct {
	processing *processingdomain.VideoProcessing
	getErr     error
	updateErr  error
	statuses   []string
}

func (f *failingRepositoryFake) GetVideoProcessingByID(string, string) (*processingdomain.VideoProcessing, error) {
	return f.processing, f.getErr
}

func (f *failingRepositoryFake) UpdateVideoProcessing(processing *processingdomain.VideoProcessing) error {
	f.statuses = append(f.statuses, processing.Status)
	return f.updateErr
}

type failingStorageFake struct {
	downloadErr error
	uploadErr   error
}

func (f *failingStorageFake) Download(context.Context, string, string, string) error {
	return f.downloadErr
}

func (f *failingStorageFake) Upload(context.Context, string, string, string, string) error {
	return f.uploadErr
}

type failingProcessorFake struct {
	err error
}

func (f failingProcessorFake) Process(context.Context, string, string) ([]string, error) {
	return nil, f.err
}

func workerMessage(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(processingdomain.VideoProcessing{ID: "video-1", BatchID: "batch-1", StoragePath: "input/video.mp4"})
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestProcessRejectsIncompleteAndMissingMessages(t *testing.T) {
	service := NewService(&failingRepositoryFake{}, &failingStorageFake{}, failingProcessorFake{}, "bucket", t.TempDir())
	for _, body := range [][]byte{[]byte(`{}`), workerMessage(t)} {
		if string(body) != `{}` {
			service = NewService(&failingRepositoryFake{}, &failingStorageFake{}, failingProcessorFake{}, "bucket", t.TempDir())
		}
		if err := service.Process(context.Background(), body); !errors.Is(err, ErrPermanent) {
			t.Fatalf("expected permanent error for %s, got %v", body, err)
		}
	}
}

func TestProcessPropagatesRepositoryAndUpdateErrors(t *testing.T) {
	message := workerMessage(t)
	lookupErr := errors.New("lookup failed")
	service := NewService(&failingRepositoryFake{getErr: lookupErr}, &failingStorageFake{}, failingProcessorFake{}, "bucket", t.TempDir())
	if err := service.Process(context.Background(), message); !errors.Is(err, lookupErr) {
		t.Fatalf("expected lookup error, got %v", err)
	}
	updateErr := errors.New("update failed")
	repository := &failingRepositoryFake{processing: &processingdomain.VideoProcessing{ID: "video-1", BatchID: "batch-1", Name: "video.mp4"}, updateErr: updateErr}
	service = NewService(repository, &failingStorageFake{}, failingProcessorFake{}, "bucket", t.TempDir())
	if err := service.Process(context.Background(), message); !errors.Is(err, updateErr) {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestProcessMarksFailedWhenProcessingFails(t *testing.T) {
	processingErr := errors.New("processor failed")
	repository := &failingRepositoryFake{processing: &processingdomain.VideoProcessing{ID: "video-1", BatchID: "batch-1", Name: "video.mp4"}}
	service := NewService(repository, &failingStorageFake{}, failingProcessorFake{err: processingErr}, "bucket", t.TempDir())
	if err := service.Process(context.Background(), workerMessage(t)); !errors.Is(err, ErrPermanent) {
		t.Fatalf("expected permanent processing error, got %v", err)
	}
	if len(repository.statuses) != 2 || repository.statuses[1] != "FAILED" {
		t.Fatalf("expected failed status transition, got %v", repository.statuses)
	}
}

func TestProcessHandlesDownloadAndUploadFailures(t *testing.T) {
	processing := &processingdomain.VideoProcessing{ID: "video-1", BatchID: "batch-1", Name: "video.mp4", OutputPath: "output/"}
	downloadErr := errors.New("download failed")
	repository := &failingRepositoryFake{processing: processing}
	service := NewService(repository, &failingStorageFake{downloadErr: downloadErr}, workerProcessorFake{}, "bucket", t.TempDir())
	if err := service.Process(context.Background(), workerMessage(t)); !errors.Is(err, ErrPermanent) {
		t.Fatalf("expected permanent download error, got %v", err)
	}

	uploadErr := errors.New("upload failed")
	service = NewService(&failingRepositoryFake{processing: processing}, &failingStorageFake{uploadErr: uploadErr}, workerProcessorFake{}, "bucket", t.TempDir())
	if err := service.Process(context.Background(), workerMessage(t)); !errors.Is(err, ErrPermanent) {
		t.Fatalf("expected permanent upload error, got %v", err)
	}
}
