package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmailAlreadyExists = errors.New("e-mail já cadastrado no sistema")

// Definimos uma interface local ou usamos diretamente a do ports na injeção do main.
// Para blindar o service de importar o pacote ports, passamos a interface de repositório por parâmetro.
type internalRepository interface {
	Create(user *User) error
	Update(user *User) error
	Delete(id string) error
	FindByEmail(email string) (*User, error)
	GetByID(id string) (*User, error)
}

type userService struct {
	repo internalRepository
}

// NewUserService retorna a struct concreta. No main.go, o Go vai aceitar
// essa struct como um ports.UserUseCase porque ela possui os métodos necessários.
func NewUserService(repo internalRepository) *userService {
	return &userService{repo: repo}
}

// Register casa perfeitamente com a assinatura exigida por ports.UserUseCase
func (s *userService) Create(cmd CreateUserCommand) (*User, error) {
	existingUser, _ := s.repo.FindByEmail(cmd.Email)
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	user := &User{
		ID:        uuid.New().String(),
		Name:      cmd.Name,
		Email:     cmd.Email,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = string(hashedPassword)

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser atualiza os dados do usuário existente
func (s *userService) Update(cmd UpdateUserCommand) (*User, error) {
	user, err := s.repo.GetByID(cmd.ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("usuário não encontrado")
	}
	if (user.Email != cmd.Email) && (cmd.Email != "") {
		existingUser, _ := s.repo.FindByEmail(cmd.Email)
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, ErrEmailAlreadyExists
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user.Name = cmd.Name
	user.Email = cmd.Email
	user.UpdatedAt = time.Now()
	user.PasswordHash = string(hashedPassword)

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser remove o usuário do sistema
func (s *userService) Delete(id string) error {
	existingUser, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if existingUser == nil {
		return errors.New("usuário não encontrado")
	}

	return s.repo.Delete(id)
}
