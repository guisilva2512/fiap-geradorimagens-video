package domain

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type userRepositoryFake struct {
	user        *User
	emailUser   *User
	createdUser *User
	users       []*User
	err         error
	updatedUser *User
	deletedID   string
}

func (f *userRepositoryFake) Create(user *User) error {
	f.createdUser = user
	return nil
}

func (f *userRepositoryFake) Update(user *User) error {
	f.updatedUser = user
	return f.err
}
func (f *userRepositoryFake) Delete(id string) error {
	f.deletedID = id
	return f.err
}
func (f *userRepositoryFake) FindByEmail(string) (*User, error) {
	if f.emailUser != nil {
		return f.emailUser, f.err
	}
	return f.user, f.err
}
func (f *userRepositoryFake) List() ([]*User, error)        { return f.users, f.err }
func (f *userRepositoryFake) GetByID(string) (*User, error) { return f.user, f.err }

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

func TestUserServiceDelegatesListAndGet(t *testing.T) {
	expected := &User{ID: "user-1"}
	repository := &userRepositoryFake{user: expected, users: []*User{expected}}
	service := NewUserService(repository)

	users, err := service.List()
	if err != nil || len(users) != 1 || users[0] != expected {
		t.Fatalf("unexpected users: %#v, %v", users, err)
	}
	user, err := service.Get(expected.ID)
	if err != nil || user != expected {
		t.Fatalf("unexpected user: %#v, %v", user, err)
	}
}

func TestUserServiceHandlesCreateAndGetErrors(t *testing.T) {
	if _, err := NewUserService(&userRepositoryFake{}).Create(CreateUserCommand{}); !errors.Is(err, ErrShortName) {
		t.Fatalf("expected invalid name error, got %v", err)
	}
	if _, err := NewUserService(&userRepositoryFake{user: &User{ID: "other"}}).Create(CreateUserCommand{Name: "Maria", Email: "maria@example.com", Password: "secret"}); !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("expected duplicate email error, got %v", err)
	}
	if _, err := NewUserService(&userRepositoryFake{}).Get("missing"); err == nil {
		t.Fatal("expected missing user error")
	}
	if _, err := NewUserService(&userRepositoryFake{err: errors.New("repository failed")}).Get("user-1"); err == nil {
		t.Fatal("expected repository error")
	}
}

func TestUserServiceUpdatesAndDeletes(t *testing.T) {
	initial := &User{ID: "user-1", Name: "Old Name", Email: "old@example.com"}
	repository := &userRepositoryFake{user: initial}
	service := NewUserService(repository)
	updated, err := service.Update(UpdateUserCommand{ID: initial.ID, Name: "New Name", Email: "new@example.com", Password: "secret"})
	if err != nil || updated.Name != "New Name" || repository.updatedUser != updated {
		t.Fatalf("unexpected update: %#v, %v", updated, err)
	}
	if err := service.Delete(initial.ID); err != nil || repository.deletedID != initial.ID {
		t.Fatalf("unexpected delete: %v", err)
	}

	repository.user = nil
	if _, err := service.Update(UpdateUserCommand{ID: "missing"}); err == nil {
		t.Fatal("expected missing user update error")
	}
	if err := service.Delete("missing"); err == nil {
		t.Fatal("expected missing user delete error")
	}
}

func TestUserServiceUpdatesWithExistingEmail(t *testing.T) {
	current := &User{ID: "user-1", Name: "Old Name", Email: "old@example.com"}
	other := &User{ID: "user-2", Email: "new@example.com"}
	repository := &userRepositoryFake{user: current}
	service := NewUserService(repository)
	if _, err := service.Update(UpdateUserCommand{ID: current.ID, Name: "New Name", Email: current.Email, Password: "secret"}); err != nil {
		t.Fatalf("expected same email update to succeed: %v", err)
	}

	repository = &userRepositoryFake{user: current, emailUser: other}
	service = NewUserService(repository)
	if _, err := service.Update(UpdateUserCommand{ID: current.ID, Email: "other@example.com", Password: "secret"}); !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("expected duplicate email error, got %v", err)
	}
}

func TestLoginAndJWTConfigurationErrors(t *testing.T) {
	service := NewUserService(&userRepositoryFake{})
	if _, _, err := service.Login(LoginUserCommand{}); err == nil {
		t.Fatal("expected missing user login error")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	service = NewUserService(&userRepositoryFake{user: &User{PasswordHash: string(hash)}})
	if _, _, err := service.Login(LoginUserCommand{Password: "wrong"}); err == nil {
		t.Fatal("expected invalid password error")
	}

	t.Setenv("JWT_SECRET", "")
	if _, err := generateJWT("user", "user"); err == nil {
		t.Fatal("expected missing secret error")
	}
	t.Setenv("JWT_SECRET", "secret")
	if _, err := generateJWT("user", "user"); err == nil {
		t.Fatal("expected missing issuer error")
	}
	t.Setenv("JWT_ISSUER", "issuer")
	if _, err := generateJWT("user", "user"); err == nil {
		t.Fatal("expected missing expiration error")
	}
	t.Setenv("JWT_EXPIRATION_HOURS", "invalid")
	if _, err := generateJWT("user", "user"); err == nil {
		t.Fatal("expected invalid expiration error")
	}
}

func TestUserValidateRejectsInvalidEmail(t *testing.T) {
	if err := (&User{Name: "Maria", Email: "x"}).Validate(); !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected invalid email error, got %v", err)
	}
}
