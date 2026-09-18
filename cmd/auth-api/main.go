package main

import (
	"log"

	"github.com/gin-gonic/gin"
	authDb "github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/adapters/db"
	authHttp "github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/adapters/http"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/domain"
	database "github.com/guisilva2512/fiap-geradorimagens-video/internal/pkg/databases"
)

func main() {
	// 1 . Inicializa a conexão com o banco de dados (Postgres) usando GORM
	db := database.Databases() // sua func de conexão do gorm

	// 2. Cria o adaptador do banco (Postgres)
	postgresRepo := authDb.NewPostgresRepository(db)

	// 3. Cria o serviço de domínio passando o repositório
	userService := domain.NewUserService(postgresRepo)

	// 4. Cria o handler injetando o serviço.
	// O Go valida implicitamente que o userService atende à porta ports.UserUseCase
	httpHandler := authHttp.NewHttpUserHandler(userService)

	// 5. Configura o roteamento do Gin e registra o endpoint de registro de usuário
	r := gin.Default()

	// Agrupamento de rotas e versionamento da API
	v1 := r.Group("/v1")
	{
		v1.GET("/users", httpHandler.ListUsers)
		v1.GET("/users/:id", httpHandler.GetUser)
		v1.POST("/users", httpHandler.CreateUser)
		v1.PUT("/users/:id", httpHandler.UpdateUser)
		v1.DELETE("/users/:id", httpHandler.DeleteUser)
		v1.POST("/login", httpHandler.LoginUser)
	}

	log.Println("🔑 Auth API (Gin) inicializada com sucesso!")

	// Inicia o servidor HTTP escutando na porta 8081
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Erro ao rodar servidor Gin: %v", err)
	}
}
