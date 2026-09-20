package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/domain"
)

type videoUseCaseFake struct {
	batches     []*domain.VideoBatch
	processings []*domain.VideoProcessing
	batch       *domain.VideoBatch
	processing  *domain.VideoProcessing
	err         error
	createUser  string
	batchID     string
	videoID     string
}

func (f *videoUseCaseFake) List() ([]*domain.VideoBatch, error)    { return f.batches, f.err }
func (f *videoUseCaseFake) Get(string) (*domain.VideoBatch, error) { return f.batch, f.err }
func (f *videoUseCaseFake) Create(c domain.CreateVideoBatchCommand) (*domain.VideoBatch, error) {
	f.createUser = c.UserID
	return f.batch, f.err
}
func (f *videoUseCaseFake) Delete(id string) error { f.batchID = id; return f.err }
func (f *videoUseCaseFake) ListProcessings(id string) ([]*domain.VideoProcessing, error) {
	f.batchID = id
	return f.processings, f.err
}
func (f *videoUseCaseFake) CreateProcessing(context.Context, domain.CreateVideoProcessingCommand) (*domain.VideoProcessing, error) {
	return f.processing, f.err
}
func (f *videoUseCaseFake) UpdateProcessing(c domain.UpdateVideoProcessingCommand) (*domain.VideoProcessing, error) {
	f.batchID, f.videoID = c.BatchID, c.ID
	return f.processing, f.err
}
func (f *videoUseCaseFake) DeleteProcessing(batchID, videoID string) error {
	f.batchID, f.videoID = batchID, videoID
	return f.err
}
func (f *videoUseCaseFake) DownloadProcessing(string, string) (io.ReadCloser, *int64, string, string, error) {
	return nil, nil, "", "", f.err
}
func (f *videoUseCaseFake) WriteImagesZip(context.Context, string, string, io.Writer) error {
	return f.err
}

func uploadContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request
	return context, recorder
}
func uploadParams(c *gin.Context, values ...string) {
	c.Params = gin.Params{{Key: "id", Value: values[0]}}
	if len(values) > 1 {
		c.Params = append(c.Params, gin.Param{Key: "video_id", Value: values[1]})
	}
}
func batchFixture() *domain.VideoBatch {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	return &domain.VideoBatch{ID: "batch-1", UserID: "user-1", CreatedAt: now, UpdatedAt: now}
}
func processingFixture() *domain.VideoProcessing {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	return &domain.VideoProcessing{ID: "video-1", BatchID: "batch-1", Status: "PENDING", Name: "clip.mp4", CreatedAt: now, UpdatedAt: now}
}

func TestVideoBatchHTTPHandlers(t *testing.T) {
	fake := &videoUseCaseFake{batches: []*domain.VideoBatch{batchFixture()}, batch: batchFixture()}
	context, recorder := uploadContext(http.MethodGet, "/uploads", "")
	handler := NewHttpVideoBatchHandler(fake)
	handler.ListUploads(context)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "2026-09-20 12:00:00") {
		t.Fatalf("unexpected list: %d %s", recorder.Code, recorder.Body.String())
	}

	context, recorder = uploadContext(http.MethodGet, "/uploads/batch-1", "")
	uploadParams(context, "batch-1")
	handler.GetUpload(context)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected get 200, got %d", recorder.Code)
	}
	context, recorder = uploadContext(http.MethodPost, "/uploads", `{`)
	handler.CreateUpload(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected malformed JSON 400, got %d", recorder.Code)
	}
	context, recorder = uploadContext(http.MethodPost, "/uploads", `{"user_id":"user-1"}`)
	handler.CreateUpload(context)
	if recorder.Code != http.StatusCreated || fake.createUser != "user-1" {
		t.Fatalf("unexpected create: %d", recorder.Code)
	}
	context, recorder = uploadContext(http.MethodDelete, "/uploads/batch-1", "")
	uploadParams(context, "batch-1")
	NewHttpVideoBatchHandler(&videoUseCaseFake{err: errors.New("failed")}).DeleteUpload(context)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected delete error 500, got %d", recorder.Code)
	}
}

func TestProcessingHTTPHandlers(t *testing.T) {
	handler := NewHttpVideoBatchHandler(&videoUseCaseFake{processings: []*domain.VideoProcessing{processingFixture()}, processing: processingFixture()})
	context, recorder := uploadContext(http.MethodGet, "/uploads//processings", "")
	handler.ListProcessings(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected missing batch 400, got %d", recorder.Code)
	}
	context, recorder = uploadContext(http.MethodGet, "/uploads/batch-1/processings", "")
	uploadParams(context, "batch-1")
	handler.ListProcessings(context)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected processing list 200, got %d", recorder.Code)
	}

	context, recorder = uploadContext(http.MethodPost, "/uploads/batch-1/processings", "")
	uploadParams(context, "batch-1")
	handler.CreateProcessing(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected missing file 400, got %d", recorder.Code)
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "clip.mp4")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("video"))
	_ = writer.Close()
	context, recorder = uploadContext(http.MethodPost, "/uploads/batch-1/processings", body.String())
	context.Request.Header.Set("Content-Type", writer.FormDataContentType())
	uploadParams(context, "batch-1")
	handler.CreateProcessing(context)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected multipart create 201, got %d %s", recorder.Code, recorder.Body.String())
	}

	context, recorder = uploadContext(http.MethodPut, "/uploads/batch-1/processings/video-1", `{"status":"DONE"}`)
	uploadParams(context, "batch-1", "video-1")
	handler.UpdateProcessing(context)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected update 200, got %d", recorder.Code)
	}
	context, recorder = uploadContext(http.MethodDelete, "/uploads/batch-1/processings/video-1", "")
	uploadParams(context, "batch-1", "video-1")
	handler.DeleteProcessing(context)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected processing delete 200, got %d", recorder.Code)
	}
}

func TestDownloadHTTPHandlers(t *testing.T) {
	for _, test := range []struct {
		name string
		path string
		call func(*HttpUserHandler, *gin.Context)
	}{
		{"batch", "/uploads/batch-1/images", func(h *HttpUserHandler, c *gin.Context) { h.DownloadBatchImages(c) }},
		{"processing", "/uploads/batch-1/processings/video-1/images", func(h *HttpUserHandler, c *gin.Context) { h.DownloadProcessingImages(c) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			context, recorder := uploadContext(http.MethodGet, test.path, "")
			uploadParams(context, "batch-1", "video-1")
			test.call(NewHttpVideoBatchHandler(&videoUseCaseFake{err: errors.New("not found")}), context)
			if recorder.Code != http.StatusNotFound || !strings.Contains(recorder.Header().Get("Content-Disposition"), "attachment") {
				t.Fatalf("unexpected download response: %d", recorder.Code)
			}
		})
	}
}
