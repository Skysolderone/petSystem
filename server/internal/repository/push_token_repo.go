package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"petverse/server/internal/model"
)

type PushTokenRepository struct {
	db *gorm.DB
}

func NewPushTokenRepository(db *gorm.DB) *PushTokenRepository {
	return &PushTokenRepository{db: db}
}

func (r *PushTokenRepository) Upsert(ctx context.Context, pushToken *model.PushToken) error {
	now := time.Now()
	if pushToken.LastSeenAt.IsZero() {
		pushToken.LastSeenAt = now
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "token"}},
			DoUpdates: clause.Assignments(map[string]any{
				"user_id":      pushToken.UserID,
				"provider":     pushToken.Provider,
				"platform":     pushToken.Platform,
				"is_active":    pushToken.IsActive,
				"last_seen_at": pushToken.LastSeenAt,
				"updated_at":   now,
				"deleted_at":   nil,
			}),
		}).
		Create(pushToken).Error
}

func (r *PushTokenRepository) ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]model.PushToken, error) {
	var items []model.PushToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("updated_at DESC").
		Find(&items).Error
	return items, err
}

func (r *PushTokenRepository) DeactivateByUserAndToken(ctx context.Context, userID uuid.UUID, token string) error {
	return r.db.WithContext(ctx).
		Model(&model.PushToken{}).
		Where("user_id = ? AND token = ?", userID, token).
		Updates(map[string]any{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error
}

func (r *PushTokenRepository) DeactivateTokens(ctx context.Context, tokens []string) error {
	if len(tokens) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.PushToken{}).
		Where("token IN ?", tokens).
		Updates(map[string]any{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error
}

func (r *PushTokenRepository) GetByUserAndToken(ctx context.Context, userID uuid.UUID, token string) (*model.PushToken, error) {
	var item model.PushToken
	err := r.db.WithContext(ctx).Where("user_id = ? AND token = ?", userID, token).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
