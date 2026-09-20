package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/domain"
)

type userUseCaseFake struct {
	users   []*domain.User
	user    *domain.User
	err     error
	created domain.CreateUserCommand
	updated domain.UpdateUserCommand
	deleted string
	got     string
	token   string
}

func (f *userUseCaseFake) List() ([]*domain.User, error)       { return f.users, f.err }
func (f *userUseCaseFake) Get(id string) (*domain.User, error) { f.got = id; return f.user, f.err }
func (f *userUseCaseFake) Create(cmd domain.CreateUserCommand) (*domain.User, error) {
	f.created = cmd
	return f.user, f.err
}
func (f *userUseCaseFake) Update(cmd domain.UpdateUserCommand) (*domain.User, error) {
	f.updated = cmd
	return f.user, f.err
}
func (f *userUseCaseFake) Delete(id string) error { f.deleted = id; return f.err }
func (f *userUseCaseFake) Login(domain.LoginUserCommand) (*domain.User, string, error) {
	return f.user, f.token, f.err
}

func userContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request
	return context, recorder
}

func TestListUsers(t *testing.T) {
	user := &domain.User{ID: "1", Name: "Maria", Email: "maria@example.com"}
	context, recorder := userContext(http.MethodGet, "/users", "")
	NewHttpUserHandler(&userUseCaseFake{users: []*domain.User{user}}).ListUsers(context)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "maria@example.com") {
		t.Fatalf("unexpected list: %d %s", recorder.Code, recorder.Body.String())
	}

	context, recorder = userContext(http.MethodGet, "/users", "")
	NewHttpUserHandler(&userUseCaseFake{err: errors.New("down")}).ListUsers(context)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

func TestGetUser(t *testing.T) {
	context, recorder := userContext(http.MethodGet, "/users/1", "")
	context.Params = gin.Params{{Key: "id", Value: "1"}}
	fake := &userUseCaseFake{user: &domain.User{ID: "1", Name: "Maria", Email: "maria@example.com"}}
	NewHttpUserHandler(fake).GetUser(context)
	if recorder.Code != http.StatusOK || fake.got != "1" {
		t.Fatalf("unexpected get: %d", recorder.Code)
	}

	context, recorder = userContext(http.MethodGet, "/users/", "")
	NewHttpUserHandler(&userUseCaseFake{}).GetUser(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected missing ID 400, got %d", recorder.Code)
	}
}

func TestCreateAndUpdateUsers(t *testing.T) {
	user := &domain.User{ID: "1", Name: "Maria", Email: "maria@example.com"}
	context, recorder := userContext(http.MethodPost, "/users", `{`)
	NewHttpUserHandler(&userUseCaseFake{}).CreateUser(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid JSON 400, got %d", recorder.Code)
	}

	context, recorder = userContext(http.MethodPost, "/users", `{"name":"M","email":"maria@example.com","password":"secret123"}`)
	NewHttpUserHandler(&userUseCaseFake{err: domain.ErrShortName}).CreateUser(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected domain error 400, got %d", recorder.Code)
	}

	context, recorder = userContext(http.MethodPost, "/users", `{"name":"Maria","email":"maria@example.com","password":"secret123"}`)
	fake := &userUseCaseFake{user: user}
	NewHttpUserHandler(fake).CreateUser(context)
	if recorder.Code != http.StatusCreated || fake.created.Email != user.Email {
		t.Fatalf("unexpected create: %d %#v", recorder.Code, fake.created)
	}

	context, recorder = userContext(http.MethodPut, "/users/1", `{"name":"Maria","email":"maria@example.com","password":"secret123"}`)
	context.Params = gin.Params{{Key: "id", Value: "1"}}
	NewHttpUserHandler(fake).UpdateUser(context)
	if recorder.Code != http.StatusOK || fake.updated.ID != "1" {
		t.Fatalf("unexpected update: %d %#v", recorder.Code, fake.updated)
	}

	context, recorder = userContext(http.MethodPut, "/users/", `{"name":"Maria","email":"maria@example.com","password":"secret123"}`)
	NewHttpUserHandler(&userUseCaseFake{}).UpdateUser(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected missing update ID 400, got %d", recorder.Code)
	}
}

func TestDeleteAndLoginUsers(t *testing.T) {
	user := &domain.User{ID: "1", Name: "Maria", Email: "maria@example.com"}
	fake := &userUseCaseFake{user: user, token: "token"}
	context, recorder := userContext(http.MethodDelete, "/users/1", "")
	context.Params = gin.Params{{Key: "id", Value: "1"}}
	NewHttpUserHandler(fake).DeleteUser(context)
	if recorder.Code != http.StatusOK || fake.deleted != "1" {
		t.Fatalf("unexpected delete: %d", recorder.Code)
	}

	context, recorder = userContext(http.MethodPost, "/login", `{"email":"maria@example.com","password":"secret123"}`)
	NewHttpUserHandler(fake).LoginUser(context)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"token":"token"`) {
		t.Fatalf("unexpected login: %d %s", recorder.Code, recorder.Body.String())
	}

	context, recorder = userContext(http.MethodPost, "/login", `{"email":"bad","password":"x"}`)
	NewHttpUserHandler(&userUseCaseFake{}).LoginUser(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid login payload 400, got %d", recorder.Code)
	}

	context, recorder = userContext(http.MethodPost, "/login", `{"email":"maria@example.com","password":"wrong"}`)
	NewHttpUserHandler(&userUseCaseFake{err: errors.New("invalid")}).LoginUser(context)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected invalid credentials 401, got %d", recorder.Code)
	}
}
