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

type GormVideoProcessing struct {
	ID           string         `gorm:"type:uuid;primaryKey"`
	BatchID      string         `gorm:"type:uuid;index"`
	Status       string         `gorm:"type:video_process_status;not null;default:'PENDING'"`
	Name         string         `gorm:"type:varchar(255);not null"`
	StoragePath  string         `gorm:"type:varchar(512);not null"`
	OutputPath   string         `gorm:"type:varchar(512)"`
	ErrorMessage string         `gorm:"type:text"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (GormVideoBatch) TableName() string {
	return "video_batches"
}

func (GormVideoProcessing) TableName() string {
	return "video_processings"
}

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresVideoBatchRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// VideoBatch methods
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

// VideoProcessing methods
func (r *PostgresRepository) ListVideoProcessings(batchId string) ([]*domain.VideoProcessing, error) {
	var dbProcessings []GormVideoProcessing
	err := r.db.Where("batch_id = ?", batchId).Find(&dbProcessings).Error
	if err != nil {
		return nil, err
	}

	domainProcessings := make([]*domain.VideoProcessing, len(dbProcessings))
	for i, dbProcessing := range dbProcessings {
		domainProcessings[i] = &domain.VideoProcessing{
			ID:           dbProcessing.ID,
			BatchID:      dbProcessing.BatchID,
			Status:       dbProcessing.Status,
			Name:         dbProcessing.Name,
			StoragePath:  dbProcessing.StoragePath,
			OutputPath:   dbProcessing.OutputPath,
			ErrorMessage: dbProcessing.ErrorMessage,
			CreatedAt:    dbProcessing.CreatedAt,
			UpdatedAt:    dbProcessing.UpdatedAt,
		}
	}

	return domainProcessings, nil
}

func (r *PostgresRepository) GetVideoProcessingByID(batchId string, id string) (*domain.VideoProcessing, error) {
	var dbProcessing GormVideoProcessing

	err := r.db.Where("batch_id = ? AND id = ?", batchId, id).First(&dbProcessing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.VideoProcessing{
		ID:           dbProcessing.ID,
		BatchID:      dbProcessing.BatchID,
		Status:       dbProcessing.Status,
		Name:         dbProcessing.Name,
		StoragePath:  dbProcessing.StoragePath,
		OutputPath:   dbProcessing.OutputPath,
		ErrorMessage: dbProcessing.ErrorMessage,
		CreatedAt:    dbProcessing.CreatedAt,
		UpdatedAt:    dbProcessing.UpdatedAt,
	}, nil
}

func (r *PostgresRepository) CreateVideoProcessing(p *domain.VideoProcessing) error {
	dbProcessing := GormVideoProcessing{
		ID:           p.ID,
		BatchID:      p.BatchID,
		Status:       p.Status,
		Name:         p.Name,
		StoragePath:  p.StoragePath,
		OutputPath:   p.OutputPath,
		ErrorMessage: p.ErrorMessage,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
	return r.db.Create(&dbProcessing).Error
}

func (r *PostgresRepository) UpdateVideoProcessing(p *domain.VideoProcessing) error {
	// var dbProcessing GormVideoProcessing
	// err := r.db.Where("id = ?", p.ID).First(&dbProcessing).Error
	// if err != nil {
	// 	return err
	// }

	// dbProcessing.BatchID = p.BatchID
	// dbProcessing.Status = p.Status
	// dbProcessing.Name = p.Name
	// dbProcessing.StoragePath = p.StoragePath
	// dbProcessing.OutputPath = p.OutputPath
	// dbProcessing.ErrorMessage = p.ErrorMessage
	// dbProcessing.CreatedAt = p.CreatedAt
	// dbProcessing.UpdatedAt = p.UpdatedAt

	// return r.db.Save(&dbProcessing).Error

	dbProcessing := GormVideoProcessing{
		ID:           p.ID,
		BatchID:      p.BatchID,
		Status:       p.Status,
		Name:         p.Name,
		StoragePath:  p.StoragePath,
		OutputPath:   p.OutputPath,
		ErrorMessage: p.ErrorMessage,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}

	return r.db.Model(&dbProcessing).Updates(dbProcessing).Error
}

func (r *PostgresRepository) DeleteVideoProcessing(batchId string, id string) error {
	var dbProcessing GormVideoProcessing

	err := r.db.Where("batch_id = ? AND id = ?", batchId, id).First(&dbProcessing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	return r.db.Delete(&dbProcessing).Error
}
