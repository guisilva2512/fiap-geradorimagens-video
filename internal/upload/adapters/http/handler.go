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

// // DownloadUpload handles the GET /uploads/:id/download endpoint to download a specific video upload.
// func (h *HttpUserHandler) DownloadUpload(c *gin.Context) {
// 	id := c.Param("id")
// 	if id == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
// 		return
// 	}

// 	videoId := c.Param("video_id")
// 	if videoId == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "video_id is required"})
// 		return
// 	}

// 	output, contentLength, contentType, _, err := h.useCase.DownloadProcessing(id, videoId)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer output.Close()

// 	key := videoId
// 	c.Header("Content-Description", "File Transfer")
// 	c.Header("Content-Disposition", "attachment; filename="+key)
// 	c.Header("Content-Type", contentType)

// 	if contentLength != nil {
// 		c.DataFromReader(http.StatusOK, *contentLength, contentType, output, nil)
// 		return
// 	}

// 	if contentType != "" {
// 		c.DataFromReader(http.StatusOK, *contentLength, contentType, output, nil)
// 		return
// 	}

// 	// Fallback streaming if Content-Length is missing
// 	c.Stream(func(w io.Writer) bool {
// 		_, streamErr := io.Copy(w, output)
// 		if streamErr != nil && !errors.Is(streamErr, io.EOF) {
// 			log.Printf("error streaming file: %v", streamErr)
// 		}
// 		return false // stop looping immediately
// 	})
// }

// // Run Processing handles the GET /uploads/:id/processings/:video_id/run endpoint to run a specific video processing.
// func (h *HttpUserHandler) RunProcessing(c *gin.Context) {
// 	id := c.Param("id")
// 	if id == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
// 		return
// 	}

// 	videoId := c.Param("video_id")
// 	if videoId == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "video_id is required"})
// 		return
// 	}

// 	output, _, _, name, err := h.useCase.DownloadProcessing(id, videoId)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer output.Close()

// 	// *****
// 	timestamp := time.Now().Format("20060102_150405")
// 	filename := fmt.Sprintf("%s_%s", timestamp, name)
// 	videoPath := filepath.Join("uploads", filename)

// 	out, err := os.Create(videoPath)
// 	if err != nil {
// 		c.JSON(500, ProcessingResult{
// 			Success: false,
// 			Message: "Erro ao salvar arquivo: " + err.Error(),
// 		})
// 		return
// 	}
// 	defer out.Close()

// 	_, err = io.Copy(out, output)
// 	if err != nil {
// 		c.JSON(500, ProcessingResult{
// 			Success: false,
// 			Message: "Erro ao salvar arquivo: " + err.Error(),
// 		})
// 		return
// 	}

// 	result := processVideo(videoPath, timestamp)

// 	if result.Success {
// 		os.Remove(videoPath)
// 	}
// 	// ****
// }

// type ProcessingResult struct {
// 	Success    bool     `json:"success"`
// 	Message    string   `json:"message"`
// 	ZipPath    string   `json:"zip_path,omitempty"`
// 	FrameCount int      `json:"frame_count,omitempty"`
// 	Images     []string `json:"images,omitempty"`
// }

// func processVideo(videoPath, timestamp string) ProcessingResult {
// 	fmt.Printf("Iniciando processamento: %s\n", videoPath)

// 	tempDir := filepath.Join("temp", timestamp)
// 	os.MkdirAll(tempDir, 0755)
// 	defer os.RemoveAll(tempDir)

// 	framePattern := filepath.Join(tempDir, "frame_%04d.png")

// 	cmd := exec.Command("ffmpeg",
// 		"-i", videoPath,
// 		"-vf", "fps=1",
// 		"-y",
// 		framePattern,
// 	)

// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return ProcessingResult{
// 			Success: false,
// 			Message: fmt.Sprintf("Erro no ffmpeg: %s\nOutput: %s", err.Error(), string(output)),
// 		}
// 	}

// 	frames, err := filepath.Glob(filepath.Join(tempDir, "*.png"))
// 	if err != nil || len(frames) == 0 {
// 		return ProcessingResult{
// 			Success: false,
// 			Message: "Nenhum frame foi extraído do vídeo",
// 		}
// 	}

// 	fmt.Printf("📸 Extraídos %d frames\n", len(frames))

// 	zipFilename := fmt.Sprintf("frames_%s.zip", timestamp)
// 	zipPath := filepath.Join("outputs", zipFilename)

// 	err = createZipFile(frames, zipPath)
// 	if err != nil {
// 		return ProcessingResult{
// 			Success: false,
// 			Message: "Erro ao criar arquivo ZIP: " + err.Error(),
// 		}
// 	}

// 	fmt.Printf("✅ ZIP criado: %s\n", zipPath)

// 	imageNames := make([]string, len(frames))
// 	for i, frame := range frames {
// 		imageNames[i] = filepath.Base(frame)
// 	}

// 	return ProcessingResult{
// 		Success:    true,
// 		Message:    fmt.Sprintf("Processamento concluído! %d frames extraídos.", len(frames)),
// 		ZipPath:    zipFilename,
// 		FrameCount: len(frames),
// 		Images:     imageNames,
// 	}
// }
// func createZipFile(files []string, zipPath string) error {
// 	zipFile, err := os.Create(zipPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer zipFile.Close()

// 	zipWriter := zip.NewWriter(zipFile)
// 	defer zipWriter.Close()

// 	for _, file := range files {
// 		err := addFileToZip(zipWriter, file)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func addFileToZip(zipWriter *zip.Writer, filename string) error {
// 	file, err := os.Open(filename)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	info, err := file.Stat()
// 	if err != nil {
// 		return err
// 	}

// 	header, err := zip.FileInfoHeader(info)
// 	if err != nil {
// 		return err
// 	}

// 	header.Name = filepath.Base(filename)
// 	header.Method = zip.Deflate

// 	writer, err := zipWriter.CreateHeader(header)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = io.Copy(writer, file)
// 	return err
// }

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
