package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/domain"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/ports"
)

//	type UserJSONRequest struct {
//		Name     string `json:"name"`
//		Email    string `json:"email"`
//		Password string `json:"password"`
//	}
//

// UserJSONRequest define o payload esperado pelo Gin
type UserJSONRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginJSONRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type HttpUserHandler struct {
	useCase ports.UserUseCase
}

func NewHttpUserHandler(uc ports.UserUseCase) *HttpUserHandler {
	return &HttpUserHandler{useCase: uc}
}

func (h *HttpUserHandler) CreateUser(c *gin.Context) {
	var req UserJSONRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos ou campos obrigatórios ausentes"})
		return
	}

	// Monta o comando mapeado dentro de domain
	cmd := domain.CreateUserCommand{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	// Executa a lógica chamando o Core (passando o contexto nativo do Go extraído do Gin)
	user, err := h.useCase.Create(cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) ||
			errors.Is(err, domain.ErrInvalidEmail) ||
			errors.Is(err, domain.ErrShortName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno no servidor"})
		return
	}

	// Resposta bem-sucedida padronizada em JSON com o status 201 Created
	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}

func (h *HttpUserHandler) UpdateUser(c *gin.Context) {
	var req UserJSONRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos ou campos obrigatórios ausentes"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do usuário é obrigatório"})
		return
	}

	// Monta o comando mapeado dentro de domain
	cmd := domain.UpdateUserCommand{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	// Executa a lógica chamando o Core (passando o contexto nativo do Go extraído do Gin)
	user, err := h.useCase.Update(cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) ||
			errors.Is(err, domain.ErrInvalidEmail) ||
			errors.Is(err, domain.ErrShortName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno no servidor"})
		return
	}

	// Resposta bem-sucedida padronizada em JSON
	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}

func (h *HttpUserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do usuário é obrigatório"})
		return
	}

	err := h.useCase.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno no servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuário excluído com sucesso"})
}

func (h *HttpUserHandler) LoginUser(c *gin.Context) {
	var req LoginJSONRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos ou campos obrigatórios ausentes"})
		return
	}

	// Monta o comando mapeado dentro de domain
	cmd := domain.LoginUserCommand{
		Email:    req.Email,
		Password: req.Password,
	}

	// Executa a lógica chamando o Core (passando o contexto nativo do Go extraído do Gin)
	user, token, err := h.useCase.Login(cmd)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciais inválidas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"token": token,
	})
}
