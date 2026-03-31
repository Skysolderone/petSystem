package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"petverse/server/internal/model"
)

type TrainingRepository struct {
	db *gorm.DB
}

func NewTrainingRepository(db *gorm.DB) *TrainingRepository {
	return &TrainingRepository{db: db}
}

func (r *TrainingRepository) Create(ctx context.Context, plan *model.TrainingPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *TrainingRepository) ListByPet(ctx context.Context, petID uuid.UUID) ([]model.TrainingPlan, error) {
	var plans []model.TrainingPlan
	err := r.db.WithContext(ctx).
		Where("pet_id = ?", petID).
		Order("created_at DESC").
		Find(&plans).Error
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *TrainingRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.TrainingPlan, error) {
	var plan model.TrainingPlan
	err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *TrainingRepository) Update(ctx context.Context, plan *model.TrainingPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *TrainingRepository) Delete(ctx context.Context, plan *model.TrainingPlan) error {
	return r.db.WithContext(ctx).Delete(plan).Error
}
