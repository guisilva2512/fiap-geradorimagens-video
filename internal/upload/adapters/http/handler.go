package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/domain"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/ports"
)

type VideoBatchJSONRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type VideoBatchJSONResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type VideoProcessingStatusUpdateRequest struct {
	Status string `json:"status" binding:"required"`
}

type VideoProcessingJSONResponse struct {
	ID           string `json:"id"`
	BatchID      string `json:"batch_id"`
	Status       string `json:"status"`
	Name         string `json:"name"`
	StoragePath  string `json:"storage_path"`
	OutputPath   string `json:"output_path"`
	ErrorMessage string `json:"error_message"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type HttpUserHandler struct {
	useCase ports.VideoBatchUseCase
}

func NewHttpVideoBatchHandler(uc ports.VideoBatchUseCase) *HttpUserHandler {
	return &HttpUserHandler{useCase: uc}
}

// VideoBatch methods
// ListUploads handles the GET /uploads endpoint to list all video uploads.
func (h *HttpUserHandler) ListUploads(c *gin.Context) {
	uploads, err := h.useCase.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert the domain objects to JSON response objects
	var jsonUploads []VideoBatchJSONResponse
	for _, u := range uploads {
		jsonUploads = append(jsonUploads, VideoBatchJSONResponse{
			ID:        u.ID,
			UserID:    u.UserID,
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: u.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": jsonUploads})
}

// GetUpload handles the GET /uploads/:id endpoint to get a specific video upload.
func (h *HttpUserHandler) GetUpload(c *gin.Context) {
	id := c.Param("id")
	upload, err := h.useCase.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert the domain object to a JSON response object
	jsonUpload := VideoBatchJSONResponse{
		ID:        upload.ID,
		UserID:    upload.UserID,
		CreatedAt: upload.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: upload.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{"data": jsonUpload})
}

// CreateUpload handles the POST /uploads endpoint to create a new video upload.
func (h *HttpUserHandler) CreateUpload(c *gin.Context) {
	var req VideoBatchJSONRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := domain.CreateVideoBatchCommand{
		UserID: req.UserID,
	}

	upload, err := h.useCase.Create(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert the domain object to a JSON response object
	jsonUpload := VideoBatchJSONResponse{
		ID:        upload.ID,
		UserID:    upload.UserID,
		CreatedAt: upload.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: upload.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	c.JSON(http.StatusCreated, gin.H{"data": jsonUpload})
}

// DeleteUpload handles the DELETE /uploads/:id endpoint to delete a specific video upload.
func (h *HttpUserHandler) DeleteUpload(c *gin.Context) {
	id := c.Param("id")
	err := h.useCase.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lote de vídeo excluído com sucesso"})
}

// Processing methods
func (h *HttpUserHandler) ListProcessings(c *gin.Context) {
	batchId := c.Param("id")
	if batchId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id is required"})
		return
	}

	processings, err := h.useCase.ListProcessings(batchId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert the domain objects to JSON response objects
	var jsonProcessings []VideoProcessingJSONResponse
	for _, p := range processings {
		jsonProcessings = append(jsonProcessings, VideoProcessingJSONResponse{
			ID:           p.ID,
			BatchID:      p.BatchID,
			Status:       p.Status,
			Name:         p.Name,
			StoragePath:  p.StoragePath,
			OutputPath:   p.OutputPath,
			ErrorMessage: p.ErrorMessage,
			CreatedAt:    p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    p.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": jsonProcessings})
}

func (h *HttpUserHandler) CreateProcessing(c *gin.Context) {
	// BatchID is passed as a URL parameter, so we don't need to bind it from the JSON request body.
	batchId := c.Param("id")
	if batchId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id is required"})
		return
	}

	// File is passed as a multipart form file, so we need to get it from the request.
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	cmd := domain.CreateVideoProcessingCommand{
		BatchID: batchId,
		File:    file,
	}

	processing, err := h.useCase.CreateProcessing(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert the domain object to a JSON response object
	jsonProcessing := VideoProcessingJSONResponse{
		ID:           processing.ID,
		BatchID:      processing.BatchID,
		Status:       processing.Status,
		Name:         processing.Name,
		StoragePath:  processing.StoragePath,
		OutputPath:   processing.OutputPath,
		ErrorMessage: processing.ErrorMessage,
		CreatedAt:    processing.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    processing.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusCreated, gin.H{"data": jsonProcessing})
}

func (h *HttpUserHandler) UpdateProcessing(c *gin.Context) {
	batchId := c.Param("id")
	if batchId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id is required"})
		return
	}

	id := c.Param("video_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video_id is required"})
		return
	}

	var req VideoProcessingStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := domain.UpdateVideoProcessingCommand{
		ID:      id,
		BatchID: batchId,
		Status:  req.Status,
	}

	processing, err := h.useCase.UpdateProcessing(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert the domain object to a JSON response object
	jsonProcessing := VideoProcessingJSONResponse{
		ID:           processing.ID,
		BatchID:      processing.BatchID,
		Status:       processing.Status,
		Name:         processing.Name,
		StoragePath:  processing.StoragePath,
		OutputPath:   processing.OutputPath,
		ErrorMessage: processing.ErrorMessage,
		CreatedAt:    processing.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    processing.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{"data": jsonProcessing})
}

func (h *HttpUserHandler) DeleteProcessing(c *gin.Context) {
	batchId := c.Param("id")
	if batchId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	videoId := c.Param("video_id")
	if videoId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video_id is required"})
		return
	}

	err := h.useCase.DeleteProcessing(batchId, videoId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Processamento de vídeo excluído com sucesso"})
}
