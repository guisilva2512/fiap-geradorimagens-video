package db

import (
	"errors"
	"time"

	"github.com/guisilva2512/fiap-geradorimagens-video/internal/auth/domain"
	"gorm.io/gorm"
)

// GormUser mapeia como o dado é persistido fisicamente na tabela 'users'
type GormUser struct {
	ID           string         `gorm:"type:uuid;primaryKey"`
	Name         string         `gorm:"type:varchar(100);not null"`
	Email        string         `gorm:"type:varchar(150);uniqueIndex;not null"`
	PasswordHash string         `gorm:"type:varchar(255);not null"`
	IsActive     bool           `gorm:"default:true;not null"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (GormUser) TableName() string {
	return "users"
}

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create implementa o contrato ports.UserRepository
func (r *PostgresRepository) Create(u *domain.User) error {
	dbUser := GormUser{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
	return r.db.Create(&dbUser).Error
}

func (r *PostgresRepository) Update(u *domain.User) error {
	dbUser := GormUser{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
	// return r.db.Save(&dbUser).Error
	return r.db.Model(&dbUser).Updates(dbUser).Error
	// return r.db.Model(&dbUser).UpdateColumns(dbUser).Error
}

// Delete implementa a deleção via ID com segurança
func (r *PostgresRepository) Delete(id string) error {
	var dbUser GormUser

	err := r.db.Where("id = ?", id).First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // Se o usuário não for encontrado, não há nada a deletar
		}
		return err // Retorna o erro se houver outro problema
	}

	// Como GormUser tem o campo DeletedAt, o GORM executará internamente um UPDATE:
	// UPDATE users SET deleted_at = '2026-09-10 ...' WHERE id = 'uuid'
	return r.db.Delete(&dbUser).Error

	// Ex sem GETID
	// err := r.db.WithContext(ctx).Delete(&GormUser{}, "id = ?", id).Error
}

// FindByEmail implementa o contrato ports.UserRepository
func (r *PostgresRepository) FindByEmail(email string) (*domain.User, error) {
	var dbUser GormUser

	err := r.db.Where("email = ?", email).First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Traduz de volta da entidade de banco para a entidade pura do Core
	return &domain.User{
		ID:           dbUser.ID,
		Name:         dbUser.Name,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		IsActive:     dbUser.IsActive,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}, nil
}

// List implementa o contrato ports.UserRepository
func (r *PostgresRepository) List() ([]*domain.User, error) {
	var dbUsers []GormUser
	err := r.db.Find(&dbUsers).Error
	if err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = &domain.User{
			ID:           dbUser.ID,
			Name:         dbUser.Name,
			Email:        dbUser.Email,
			PasswordHash: dbUser.PasswordHash,
			IsActive:     dbUser.IsActive,
			CreatedAt:    dbUser.CreatedAt,
			UpdatedAt:    dbUser.UpdatedAt,
		}
	}

	return users, nil
}

// GetByID implementa o contrato ports.UserRepository
func (r *PostgresRepository) GetByID(id string) (*domain.User, error) {
	var dbUser GormUser

	err := r.db.Where("id = ?", id).First(&dbUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.User{
		ID:           dbUser.ID,
		Name:         dbUser.Name,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		IsActive:     dbUser.IsActive,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}, nil
}
