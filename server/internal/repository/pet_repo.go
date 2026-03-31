package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"petverse/server/internal/model"
	"petverse/server/internal/pkg/pagination"
)

type PetRepository struct {
	db *gorm.DB
}

func NewPetRepository(db *gorm.DB) *PetRepository {
	return &PetRepository{db: db}
}

func (r *PetRepository) Create(ctx context.Context, pet *model.Pet) error {
	return r.db.WithContext(ctx).Create(pet).Error
}

func (r *PetRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]model.Pet, int64, error) {
	var (
		pets  []model.Pet
		total int64
	)

	tx := r.db.WithContext(ctx).Model(&model.Pet{}).Where("owner_id = ?", ownerID)
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(pagination.Offset(page, pageSize)).
		Find(&pets).
		Error
	if err != nil {
		return nil, 0, err
	}

	return pets, total, nil
}

func (r *PetRepository) GetByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*model.Pet, error) {
	var pet model.Pet
	err := r.db.WithContext(ctx).Where("id = ? AND owner_id = ?", id, ownerID).First(&pet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pet, nil
}

func (r *PetRepository) Update(ctx context.Context, pet *model.Pet) error {
	return r.db.WithContext(ctx).Save(pet).Error
}

func (r *PetRepository) Delete(ctx context.Context, pet *model.Pet) error {
	return r.db.WithContext(ctx).Delete(pet).Error
}
