package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidEmail = errors.New("formato de e-mail inválido")
	ErrShortName    = errors.New("o nome deve ter pelo menos 3 caracteres")
)

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) Validate() error {
	if len(strings.TrimSpace(u.Name)) < 3 {
		return ErrShortName
	}
	if !strings.Contains(u.Email, "@") || len(u.Email) < 5 {
		return ErrInvalidEmail
	}
	return nil
}

// CreateUserCommand colocamos a struct de comando aqui dentro do domínio.
// Assim, o service a consome nativamente sem precisar ir buscar em 'ports'.
type CreateUserCommand struct {
	Name     string
	Email    string
	Password string
}

// UpdateUserCommand colocamos a struct de comando aqui dentro do domínio.
// Assim, o service a consome nativamente sem precisar ir buscar em 'ports'.
type UpdateUserCommand struct {
	ID       string
	Name     string
	Email    string
	Password string
}

// LoginUserCommand colocamos a struct de comando aqui dentro do domínio.
// Assim, o service a consome nativamente sem precisar ir buscar em 'ports'.
type LoginUserCommand struct {
	Email    string
	Password string
}
