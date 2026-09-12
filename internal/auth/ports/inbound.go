package ports

import (
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/domain"
)

// UserUseCase agora usa o comando que reside dentro do pacote domain
// Esta interface é usado no service.go
type UserUseCase interface {
	Create(cmd domain.CreateUserCommand) (*domain.User, error)
	Update(cmd domain.UpdateUserCommand) (*domain.User, error)
	Delete(id string) error
	Login(cmd domain.LoginUserCommand) (*domain.User, string, error)
}
