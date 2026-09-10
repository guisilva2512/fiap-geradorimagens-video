package ports

import "github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/domain"

type UserRepository interface {
	Create(user *domain.CreateUserCommand) error
	Update(user *domain.UpdateUserCommand) error
	Delete(id string) error
	FindByEmail(email string) (*domain.User, error)
	GetByID(id string) (*domain.User, error)
}
