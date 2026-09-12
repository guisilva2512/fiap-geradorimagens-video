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

type HttpUserHandler struct {
	useCase ports.VideoBatchUseCase
}

func NewHttpVideoBatchHandler(uc ports.VideoBatchUseCase) *HttpUserHandler {
	return &HttpUserHandler{useCase: uc}
}

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
