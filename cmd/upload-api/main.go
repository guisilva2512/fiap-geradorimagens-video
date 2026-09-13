package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	database "github.com/guisilva2512/fiap-geradorimagens-video/internal/pkg/databases"
	videoBatchDb "github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/adapters/db"
	videoBatchHttp "github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/adapters/http"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/adapters/storage"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/domain"
)

func main() {
	repository := repository()
	awsS3Storage := aWSStorage()
	videoBatchService := domain.NewVideoBatchService(repository, awsS3Storage)
	httpHandler := videoBatchHttp.NewHttpVideoBatchHandler(videoBatchService)

	r := server(httpHandler)

	if err := r.Run(":8082"); err != nil {
		log.Fatalf("Erro ao rodar servidor Gin: %v", err)
	}
}

func repository() domain.UploadRepository {
	db := database.Databases()
	postgresRepo := videoBatchDb.NewPostgresVideoBatchRepository(db)
	return postgresRepo
}

func aWSStorage() domain.UploadStorage {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	endpoint := os.Getenv("AWS_S3_ENDPOINT")
	region := os.Getenv("AWS_REGION")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return storage.NewAWSS3StorageAdapter(client)
}

func server(httpHandler *videoBatchHttp.HttpUserHandler) *gin.Engine {
	r := gin.Default()

	// Agrupamento de rotas e versionamento da API
	v1 := r.Group("/v1")
	{
		// VideoBatch endpoints
		v1.GET("/uploads", httpHandler.ListUploads)
		v1.GET("/uploads/:id", httpHandler.GetUpload)
		v1.POST("/uploads", httpHandler.CreateUpload)
		v1.DELETE("/uploads/:id", httpHandler.DeleteUpload)

		// VideoProcessing endpoints
		v1.GET("/uploads/:id/processings", httpHandler.ListProcessings)
		v1.POST("/uploads/:id/processings", httpHandler.CreateProcessing)
		v1.PUT("/uploads/:id/processings/:video_id", httpHandler.UpdateProcessing)
		v1.DELETE("/uploads/:id/processings/:video_id", httpHandler.DeleteProcessing)
	}

	log.Println("🔑 Auth API (Gin) inicializada com sucesso!")

	return r
}
