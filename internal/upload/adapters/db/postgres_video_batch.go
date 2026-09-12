package db

import (
	"errors"
	"time"

	"github.com/guisilva2512/fiap-geradorimagens-video/internal/upload/domain"
	"gorm.io/gorm"
)

type GormVideoBatch struct {
	ID        string         `gorm:"type:uuid;primaryKey"`
	UserID    string         `gorm:"type:uuid;not null"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (GormVideoBatch) TableName() string {
	return "video_batches"
}

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresVideoBatchRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List() ([]*domain.VideoBatch, error) {
	var dbBatches []GormVideoBatch
	err := r.db.Find(&dbBatches).Error
	if err != nil {
		return nil, err
	}

	domainBatches := make([]*domain.VideoBatch, len(dbBatches))
	for i, dbBatch := range dbBatches {
		domainBatches[i] = &domain.VideoBatch{
			ID:        dbBatch.ID,
			UserID:    dbBatch.UserID,
			CreatedAt: dbBatch.CreatedAt,
			UpdatedAt: dbBatch.UpdatedAt,
		}
	}

	return domainBatches, nil
}

func (r *PostgresRepository) GetByID(id string) (*domain.VideoBatch, error) {
	var dbBatch GormVideoBatch

	err := r.db.Where("id = ?", id).First(&dbBatch).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.VideoBatch{
		ID:        dbBatch.ID,
		UserID:    dbBatch.UserID,
		CreatedAt: dbBatch.CreatedAt,
		UpdatedAt: dbBatch.UpdatedAt,
	}, nil
}

func (r *PostgresRepository) Create(b *domain.VideoBatch) error {
	dbUser := GormVideoBatch{
		ID:        b.ID,
		UserID:    b.UserID,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
	return r.db.Create(&dbUser).Error
}

func (r *PostgresRepository) Delete(id string) error {
	var dbUser GormVideoBatch

	err := r.db.Where("id = ?", id).First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	return r.db.Delete(&dbUser).Error
}
