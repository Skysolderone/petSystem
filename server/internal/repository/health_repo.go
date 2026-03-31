package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"petverse/server/internal/model"
	"petverse/server/internal/pkg/pagination"
)

type HealthRepository struct {
	db *gorm.DB
}

func NewHealthRepository(db *gorm.DB) *HealthRepository {
	return &HealthRepository{db: db}
}

func (r *HealthRepository) CreateRecord(ctx context.Context, record *model.HealthRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *HealthRepository) ListRecordsByPet(ctx context.Context, petID uuid.UUID, page, pageSize int, typeFilter string) ([]model.HealthRecord, int64, error) {
	var (
		records []model.HealthRecord
		total   int64
	)

	query := r.db.WithContext(ctx).Model(&model.HealthRecord{}).Where("pet_id = ?", petID)
	if typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("recorded_at DESC").
		Limit(pageSize).
		Offset(pagination.Offset(page, pageSize)).
		Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *HealthRepository) ListRecentRecordsByPet(ctx context.Context, petID uuid.UUID, limit int) ([]model.HealthRecord, error) {
	var records []model.HealthRecord
	err := r.db.WithContext(ctx).
		Where("pet_id = ?", petID).
		Order("recorded_at DESC").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (r *HealthRepository) GetRecordByID(ctx context.Context, petID, id uuid.UUID) (*model.HealthRecord, error) {
	var record model.HealthRecord
	err := r.db.WithContext(ctx).Where("pet_id = ? AND id = ?", petID, id).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *HealthRepository) UpdateRecord(ctx context.Context, record *model.HealthRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *HealthRepository) DeleteRecord(ctx context.Context, record *model.HealthRecord) error {
	return r.db.WithContext(ctx).Delete(record).Error
}

func (r *HealthRepository) CreateAlert(ctx context.Context, alert *model.HealthAlert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *HealthRepository) ListAlertsByPet(ctx context.Context, petID uuid.UUID) ([]model.HealthAlert, error) {
	var alerts []model.HealthAlert
	err := r.db.WithContext(ctx).
		Where("pet_id = ?", petID).
		Order("created_at DESC").
		Limit(20).
		Find(&alerts).Error
	if err != nil {
		return nil, err
	}
	return alerts, nil
}
