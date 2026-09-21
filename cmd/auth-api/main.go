package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	authDb "github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/adapters/db"
	authHttp "github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/adapters/http"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/domain"
	database "github.com/guisilva2512/fiap-geradorimagens-video/internal/pkg/databases"
	commonmiddleware "github.com/guisilva2512/fiap-geradorimagens-video/internal/pkg/middleware"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/pkg/observability"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	shutdownTracer, err := observability.InitTracer(context.Background(), "auth-api")
	if err != nil {
		log.Fatalf("inicializar OpenTelemetry: %v", err)
	}
	defer shutdownTracer(context.Background())

	// 1 . Inicializa a conexão com o banco de dados (Postgres) usando GORM
	db := database.Databases() // sua func de conexão do gorm

	// 2. Cria o adaptador do banco (Postgres)
	postgresRepo := authDb.NewPostgresRepository(db)

	// 3. Cria o serviço de domínio passando o repositório
	userService := domain.NewUserService(postgresRepo)

	// 4. Cria o handler injetando o serviço.
	// O Go valida implicitamente que o userService atende à porta ports.UserUseCase
	httpHandler := authHttp.NewHttpUserHandler(userService)

	r := server(httpHandler)

	log.Println("🔑 Auth API (Gin) inicializada com sucesso!")

	// Inicia o servidor HTTP escutando na porta 8081
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Erro ao rodar servidor Gin: %v", err)
	}
}

func server(httpHandler *authHttp.HttpUserHandler) *gin.Engine {
	r := gin.Default()
	r.Use(otelgin.Middleware("auth-api"), observability.MetricsMiddleware("auth-api"))
	r.GET("/metrics", gin.WrapH(observability.MetricsHandler()))

	v1 := r.Group("/v1")
	{
		v1.POST("/users", httpHandler.CreateUser)
		v1.POST("/login", httpHandler.LoginUser)

		protected := v1.Group("", commonmiddleware.AuthMiddleware())
		protected.GET("/users", httpHandler.ListUsers)
		protected.GET("/users/:id", httpHandler.GetUser)
		protected.PUT("/users/:id", httpHandler.UpdateUser)
		protected.DELETE("/users/:id", httpHandler.DeleteUser)
	}

	return r
}
