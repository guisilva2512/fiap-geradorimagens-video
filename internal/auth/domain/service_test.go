package domain

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type userRepositoryFake struct {
	user        *User
	createdUser *User
}

func (f *userRepositoryFake) Create(user *User) error {
	f.createdUser = user
	return nil
}

func (f *userRepositoryFake) Update(*User) error                { return nil }
func (f *userRepositoryFake) Delete(string) error               { return nil }
func (f *userRepositoryFake) FindByEmail(string) (*User, error) { return f.user, nil }
func (f *userRepositoryFake) List() ([]*User, error)            { return nil, nil }
func (f *userRepositoryFake) GetByID(string) (*User, error)     { return f.user, nil }

func TestCreateUserHashesPassword(t *testing.T) {
	repository := &userRepositoryFake{}
	user, err := NewUserService(repository).Create(CreateUserCommand{
		Name: "Maria Silva", Email: "maria@example.com", Password: "secret123",
	})

	if err != nil {
		t.Fatalf("expected user creation to succeed: %v", err)
	}
	if user.ID == "" || repository.createdUser == nil {
		t.Fatal("expected created user with an ID")
	}
	if user.PasswordHash == "secret123" {
		t.Fatal("expected password to be hashed")
	}
}

func TestLoginReturnsJWTForValidPassword(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_ISSUER", "test-issuer")
	t.Setenv("JWT_EXPIRATION_HOURS", "1")

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to prepare password hash: %v", err)
	}
	repository := &userRepositoryFake{user: &User{
		ID: "user-1", Name: "Maria Silva", Email: "maria@example.com",
		PasswordHash: string(hash), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}}

	user, token, err := NewUserService(repository).Login(LoginUserCommand{
		Email: "maria@example.com", Password: "secret123",
	})

	if err != nil {
		t.Fatalf("expected login to succeed: %v", err)
	}
	if user.ID != "user-1" || token == "" {
		t.Fatal("expected user and JWT token")
	}
}
