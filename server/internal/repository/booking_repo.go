package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"petverse/server/internal/model"
	"petverse/server/internal/pkg/pagination"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(ctx context.Context, booking *model.Booking) error {
	return r.db.WithContext(ctx).Create(booking).Error
}

func (r *BookingRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Booking, int64, error) {
	var (
		bookings []model.Booking
		total    int64
	)

	query := r.db.WithContext(ctx).Model(&model.Booking{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("start_time DESC").
		Limit(pageSize).
		Offset(pagination.Offset(page, pageSize)).
		Find(&bookings).Error
	if err != nil {
		return nil, 0, err
	}
	return bookings, total, nil
}

func (r *BookingRepository) GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*model.Booking, error) {
	var booking model.Booking
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&booking).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) Update(ctx context.Context, booking *model.Booking) error {
	return r.db.WithContext(ctx).Save(booking).Error
}
