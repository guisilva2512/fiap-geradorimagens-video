package domain

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsValidVideoFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		valid    bool
	}{
		{name: "mp4", filename: "video.mp4", valid: true},
		{name: "uppercase extension", filename: "video.MKV", valid: true},
		{name: "unsupported extension", filename: "video.txt", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidVideoFile(test.filename); got != test.valid {
				t.Fatalf("isValidVideoFile(%q) = %v, want %v", test.filename, got, test.valid)
			}
		})
	}
}

type repositoryFake struct {
	batch             *VideoBatch
	processing        *VideoProcessing
	processings       []*VideoProcessing
	createdBatch      *VideoBatch
	createdProcessing *VideoProcessing
	updatedProcessing *VideoProcessing
	deletedBatch      string
	deletedProcessing string
	err               error
}

func (f *repositoryFake) List() ([]*VideoBatch, error)        { return nil, f.err }
func (f *repositoryFake) GetByID(string) (*VideoBatch, error) { return f.batch, f.err }
func (f *repositoryFake) Create(v *VideoBatch) error          { f.createdBatch = v; return f.err }
func (f *repositoryFake) Delete(id string) error              { f.deletedBatch = id; return f.err }
func (f *repositoryFake) ListVideoProcessings(string) ([]*VideoProcessing, error) {
	return f.processings, f.err
}
func (f *repositoryFake) GetVideoProcessingByID(string, string) (*VideoProcessing, error) {
	return f.processing, f.err
}
func (f *repositoryFake) CreateVideoProcessing(v *VideoProcessing) error {
	f.createdProcessing = v
	return f.err
}
func (f *repositoryFake) UpdateVideoProcessing(v *VideoProcessing) error {
	f.updatedProcessing = v
	return f.err
}
func (f *repositoryFake) DeleteVideoProcessing(_, id string) error {
	f.deletedProcessing = id
	return f.err
}

type storageFake struct {
	keys     []string
	files    map[string]string
	savedKey string
	output   io.ReadCloser
	length   *int64
	typeName string
	err      error
}

func (f *storageFake) SaveFile(_, key string, file multipart.File) error {
	f.savedKey = key
	_, _ = io.Copy(io.Discard, file)
	return f.err
}
func (f *storageFake) GetFile(string, string) (io.ReadCloser, *int64, string, error) {
	return f.output, f.length, f.typeName, f.err
}
func (f *storageFake) GetFileContext(_ context.Context, _, key string) (io.ReadCloser, *int64, string, error) {
	if f.err != nil {
		return nil, nil, "", f.err
	}
	return io.NopCloser(strings.NewReader(f.files[key])), nil, "image/png", nil
}
func (f *storageFake) List(context.Context, string, string) ([]string, error) { return f.keys, f.err }

type publisherFake struct {
	body []byte
	err  error
}

func (f *publisherFake) PublishMessage(_ context.Context, body []byte) error {
	f.body = body
	return f.err
}

func videoFileHeader(t *testing.T, name string) *multipart.FileHeader {
	t.Helper()
	request := httptest.NewRequest("POST", "/", strings.NewReader("--boundary\r\nContent-Disposition: form-data; name=\"file\"; filename=\""+name+"\"\r\nContent-Type: video/mp4\r\n\r\nvideo\r\n--boundary--\r\n"))
	request.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	if err := request.ParseMultipartForm(1024); err != nil {
		t.Fatal(err)
	}
	return request.MultipartForm.File["file"][0]
}

func TestVideoBatchCRUD(t *testing.T) {
	repository := &repositoryFake{batch: &VideoBatch{ID: "batch-1"}}
	service := NewVideoBatchService(repository, &storageFake{}, "bucket", &publisherFake{})
	created, err := service.Create(CreateVideoBatchCommand{UserID: "user-1"})
	if err != nil || created.ID == "" || repository.createdBatch.UserID != "user-1" {
		t.Fatalf("unexpected create: %#v %v", created, err)
	}
	if _, err = service.Get("batch-1"); err != nil {
		t.Fatal(err)
	}
	if err = service.Delete("batch-1"); err != nil || repository.deletedBatch != "batch-1" {
		t.Fatalf("unexpected delete: %v", err)
	}
	repository.batch = nil
	if _, err = service.Get("missing"); err == nil {
		t.Fatal("expected missing batch error")
	}
	if err = service.Delete("missing"); err == nil {
		t.Fatal("expected missing batch delete error")
	}
}

func TestCreateProcessingCoordinatesDependencies(t *testing.T) {
	repository := &repositoryFake{}
	storage := &storageFake{}
	publisher := &publisherFake{}
	service := NewVideoBatchService(repository, storage, "bucket", publisher)
	created, err := service.CreateProcessing(context.Background(), CreateVideoProcessingCommand{BatchID: "batch-1", File: videoFileHeader(t, "clip.MP4")})
	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if created.Status != "PENDING" || repository.createdProcessing != created || storage.savedKey == "" || len(publisher.body) == 0 {
		t.Fatalf("dependencies not coordinated: %#v", created)
	}
	_, err = service.CreateProcessing(context.Background(), CreateVideoProcessingCommand{BatchID: "batch-1", File: videoFileHeader(t, "notes.txt")})
	if err == nil || err.Error() != "invalid video file format" {
		t.Fatalf("expected invalid extension error, got %v", err)
	}
}

func TestProcessingUpdateDeleteAndZip(t *testing.T) {
	processing := &VideoProcessing{ID: "video-1", BatchID: "batch-1", Status: "PENDING", OutputPath: "output/"}
	repository := &repositoryFake{processing: processing}
	service := NewVideoBatchService(repository, &storageFake{}, "bucket", &publisherFake{})
	updated, err := service.UpdateProcessing(UpdateVideoProcessingCommand{ID: "video-1", BatchID: "batch-1", Status: "DONE"})
	if err != nil || updated.Status != "DONE" || repository.updatedProcessing != updated {
		t.Fatalf("unexpected update: %#v %v", updated, err)
	}
	if err = service.DeleteProcessing("batch-1", "video-1"); err != nil || repository.deletedProcessing != "video-1" {
		t.Fatalf("unexpected delete: %v", err)
	}

	storage := &storageFake{keys: []string{"output/z.txt", "output/b.PNG", "output/a.jpg"}, files: map[string]string{"output/b.PNG": "b", "output/a.jpg": "a"}}
	repository.processing = processing
	service = NewVideoBatchService(repository, storage, "bucket", &publisherFake{})
	var output bytes.Buffer
	if err = service.WriteImagesZip(context.Background(), "batch-1", "video-1", &output); err != nil {
		t.Fatalf("expected zip success, got %v", err)
	}
	archive, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil || len(archive.File) != 2 || archive.File[0].Name != "video-1/a.jpg" || archive.File[1].Name != "video-1/b.PNG" {
		t.Fatalf("unexpected archive: %v", err)
	}
}

func TestWriteImagesZipErrors(t *testing.T) {
	var output bytes.Buffer
	service := NewVideoBatchService(&repositoryFake{}, &storageFake{}, "bucket", &publisherFake{})
	if err := service.WriteImagesZip(context.Background(), "batch-1", "video-1", &output); err == nil {
		t.Fatal("expected missing processing error")
	}
	processing := &VideoProcessing{ID: "video-1", OutputPath: "output/"}
	service = NewVideoBatchService(&repositoryFake{processing: processing}, &storageFake{keys: []string{"output/readme.txt"}}, "bucket", &publisherFake{})
	if err := service.WriteImagesZip(context.Background(), "batch-1", "video-1", &output); err == nil {
		t.Fatal("expected no images error")
	}
	service = NewVideoBatchService(&repositoryFake{processing: processing}, &storageFake{err: errors.New("storage failed")}, "bucket", &publisherFake{})
	if err := service.WriteImagesZip(context.Background(), "batch-1", "video-1", &output); err == nil {
		t.Fatal("expected storage error")
	}
}

func TestVideoBatchListAndProcessingList(t *testing.T) {
	batch := &VideoBatch{ID: "batch-1"}
	processing := &VideoProcessing{ID: "video-1"}
	repository := &repositoryFake{batch: batch, processings: []*VideoProcessing{processing}}
	service := NewVideoBatchService(repository, &storageFake{}, "bucket", &publisherFake{})

	if batches, err := service.List(); err != nil || len(batches) != 0 {
		t.Fatalf("unexpected batches: %#v, %v", batches, err)
	}
	if processings, err := service.ListProcessings(batch.ID); err != nil || len(processings) != 1 || processings[0] != processing {
		t.Fatalf("unexpected processings: %#v, %v", processings, err)
	}
}

func TestDownloadProcessingReturnsFileMetadata(t *testing.T) {
	processing := &VideoProcessing{ID: "video-1", Name: "clip.mp4", StoragePath: "input/clip.mp4"}
	length := int64(42)
	storage := &storageFake{output: io.NopCloser(strings.NewReader("video")), length: &length, typeName: "video/mp4"}
	service := NewVideoBatchService(&repositoryFake{processing: processing}, storage, "bucket", &publisherFake{})

	file, gotLength, contentType, name, err := service.DownloadProcessing("batch-1", processing.ID)
	if err != nil || file == nil || gotLength != &length || contentType != "video/mp4" || name != "clip.mp4" {
		t.Fatalf("unexpected download result: %v, %v, %q, %q", err, gotLength, contentType, name)
	}
	file.Close()
}

func TestCreateProcessingPropagatesDependencyErrors(t *testing.T) {
	dependencyError := errors.New("dependency failed")
	file := videoFileHeader(t, "clip.mp4")

	service := NewVideoBatchService(&repositoryFake{}, &storageFake{err: dependencyError}, "bucket", &publisherFake{})
	if _, err := service.CreateProcessing(context.Background(), CreateVideoProcessingCommand{BatchID: "batch-1", File: file}); !errors.Is(err, dependencyError) {
		t.Fatalf("expected storage error, got %v", err)
	}
	service = NewVideoBatchService(&repositoryFake{err: dependencyError}, &storageFake{}, "bucket", &publisherFake{})
	if _, err := service.CreateProcessing(context.Background(), CreateVideoProcessingCommand{BatchID: "batch-1", File: file}); !errors.Is(err, dependencyError) {
		t.Fatalf("expected repository error, got %v", err)
	}
	service = NewVideoBatchService(&repositoryFake{}, &storageFake{}, "bucket", &publisherFake{err: dependencyError})
	if _, err := service.CreateProcessing(context.Background(), CreateVideoProcessingCommand{BatchID: "batch-1", File: file}); !errors.Is(err, dependencyError) {
		t.Fatalf("expected publisher error, got %v", err)
	}
}

func TestProcessingAndDownloadErrors(t *testing.T) {
	dependencyError := errors.New("dependency failed")
	service := NewVideoBatchService(&repositoryFake{err: dependencyError}, &storageFake{}, "bucket", &publisherFake{})
	if _, err := service.UpdateProcessing(UpdateVideoProcessingCommand{ID: "video-1"}); !errors.Is(err, dependencyError) {
		t.Fatalf("expected update lookup error, got %v", err)
	}
	if _, _, _, _, err := service.DownloadProcessing("batch-1", "video-1"); !errors.Is(err, dependencyError) {
		t.Fatalf("expected download lookup error, got %v", err)
	}
	if err := service.DeleteProcessing("batch-1", "video-1"); !errors.Is(err, dependencyError) {
		t.Fatalf("expected delete lookup error, got %v", err)
	}

	processing := &VideoProcessing{ID: "video-1", Name: "clip.mp4", StoragePath: "input/clip.mp4"}
	service = NewVideoBatchService(&repositoryFake{processing: processing}, &storageFake{err: dependencyError}, "bucket", &publisherFake{})
	if _, _, _, _, err := service.DownloadProcessing("batch-1", "video-1"); !errors.Is(err, dependencyError) {
		t.Fatalf("expected download storage error, got %v", err)
	}
}

func TestWriteImagesZipForAllProcessings(t *testing.T) {
	processings := []*VideoProcessing{{ID: "video-1", OutputPath: "output/one/"}, {ID: "video-2", OutputPath: "output/two/"}}
	storage := &storageFake{keys: []string{"frame.png"}, files: map[string]string{"frame.png": "image"}}
	service := NewVideoBatchService(&repositoryFake{processings: processings}, storage, "bucket", &publisherFake{})
	var output bytes.Buffer
	if err := service.WriteImagesZip(context.Background(), "batch-1", "", &output); err != nil {
		t.Fatalf("expected zip success, got %v", err)
	}
	archive, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil || len(archive.File) != 2 {
		t.Fatalf("unexpected archive: %v", err)
	}
}
