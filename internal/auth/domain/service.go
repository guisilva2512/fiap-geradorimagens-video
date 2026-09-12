package domain

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	List() ([]*User, error)
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

// ListUsers retorna todos os usuários do sistema
func (s *userService) List() ([]*User, error) {
	return s.repo.List()
}

// GetUsers retorna um usuário específico pelo ID
func (s *userService) Get(id string) (*User, error) {
	existingUser, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if existingUser == nil {
		return nil, errors.New("usuário não encontrado")
	}

	return s.repo.GetByID(id)
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

// Login autentica o usuário com base no e-mail e senha fornecidos
func (s *userService) Login(cmd LoginUserCommand) (*User, string, error) {
	user, err := s.repo.FindByEmail(cmd.Email)
	if err != nil {
		return nil, "", err
	}

	if user == nil {
		return nil, "", errors.New("usuário não encontrado")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.Password)); err != nil {
		return nil, "", errors.New("senha incorreta")
	}

	token, err := generateJWT(user.Name, "user")
	if err != nil {
		return nil, "", errors.New("falha ao gerar token JWT")
	}

	return user, token, nil
}

// CustomClaims defines the structure for data stored inside the token payload
type CustomClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func generateJWT(username, role string) (string, error) {
	// Read the secret key and other JWT settings from environment variables
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", errors.New("JWT_SECRET não definido no ambiente")
	}

	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		return "", errors.New("JWT_ISSUER não definido no ambiente")
	}

	expiration := os.Getenv("JWT_EXPIRATION_HOURS")
	if expiration == "" {
		return "", errors.New("JWT_EXPIRATION_HOURS não definido no ambiente")
	}

	expirationHours, err := strconv.Atoi(expiration)
	if err != nil {
		return "", errors.New("JWT_EXPIRATION_HOURS deve ser um número")
	}

	// Define your secret key (keep this safe, ideally in environment variables)
	var jwtKey = []byte(secretKey)

	// 1. Set up the token expiration time (e.g., 2 hours)
	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)

	// 2. Populate the custom and standard claims
	claims := &CustomClaims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	}

	// 3. Create the token using the HS256 algorithm and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 4. Sign the token with the secret key to get the final string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
