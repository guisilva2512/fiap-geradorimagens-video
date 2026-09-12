package main

import (
	"log"

	"github.com/gin-gonic/gin"
	database "github.com/guisilva2512/fiap-geradorimagens-video/internal/pkg/databases"
	videoBatchDb "github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/adapters/db"
	videoBatchHttp "github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/adapters/http"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/domain"
)

func main() {
	// 1 . Inicializa a conexão com o banco de dados (Postgres) usando GORM
	db := database.InitGorm() // sua func de conexão do gorm

	// 2. Cria o adaptador do banco (Postgres)
	postgresRepo := videoBatchDb.NewPostgresVideoBatchRepository(db)

	// 3. Cria o serviço de domínio passando o repositório
	videoBatchService := domain.NewVideoBatchService(postgresRepo)

	// 4. Cria o handler injetando o serviço.
	// O Go valida implicitamente que o userService atende à porta ports.VideoBatchUseCase
	httpHandler := videoBatchHttp.NewHttpVideoBatchHandler(videoBatchService)

	// 5. Configura o roteamento do Gin e registra o endpoint de registro de usuário
	r := gin.Default()

	// Agrupamento de rotas e versionamento da API
	v1 := r.Group("/v1")
	{
		v1.GET("/uploads", httpHandler.ListUploads)
		v1.GET("/uploads/:id", httpHandler.GetUpload)
		v1.POST("/uploads", httpHandler.CreateUpload)
		v1.DELETE("/uploads/:id", httpHandler.DeleteUpload)
	}

	log.Println("🔑 Auth API (Gin) inicializada com sucesso!")

	// Inicia o servidor HTTP escutando na porta 8082
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("Erro ao rodar servidor Gin: %v", err)
	}
}
