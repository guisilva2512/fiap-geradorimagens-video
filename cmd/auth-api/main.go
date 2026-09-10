package main

import (
	"log"

	"github.com/gin-gonic/gin"
	authDb "github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/adapters/db"
	authHttp "github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/adapters/http"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/domain"
)

func main() {
	// // Opcional: Define o modo do Gin baseado em variáveis de ambiente (debug, release, test)
	// if os.Getenv("GO_ENV") == "production" {
	// 	gin.SetMode(gin.ReleaseMode)
	// }

	// 1 . Inicializa a conexão com o banco de dados (Postgres) usando GORM
	db := authDb.InitGorm() // sua func de conexão do gorm

	// 2. Cria o adaptador do banco (Postgres)
	postgresRepo := authDb.NewPostgresRepository(db)

	// 3. Cria o serviço de domínio passando o repositório
	userService := domain.NewUserService(postgresRepo)

	// 4. Cria o handler injetando o serviço.
	// O Go valida implicitamente que o userService atende à porta ports.UserUseCase
	httpHandler := authHttp.NewHttpUserHandler(userService)

	// 5. Configura o roteamento do Gin e registra o endpoint de registro de usuário
	r := gin.Default()
	// r.POST("/v1/users", httpHandler.RegisterUser)
	// r.Run("127.0.0.1:8081")

	// Agrupamento de rotas e versionamento da API
	v1 := r.Group("/v1")
	{
		v1.POST("/users", httpHandler.CreateUser)
		v1.PUT("/users/:id", httpHandler.UpdateUser)
		v1.DELETE("/users/:id", httpHandler.DeleteUser)
	}

	log.Println("🔑 Auth API (Gin) inicializada com sucesso!")

	// Inicia o servidor HTTP escutando na porta 8081
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Erro ao rodar servidor Gin: %v", err)
	}
}
