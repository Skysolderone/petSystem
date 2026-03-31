package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
)

type ServiceProviderRepository struct {
	db *gorm.DB
}

func NewServiceProviderRepository(db *gorm.DB) *ServiceProviderRepository {
	return &ServiceProviderRepository{db: db}
}

func (r *ServiceProviderRepository) EnsureDemoProviders(ctx context.Context) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ServiceProvider{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seedUsers := []model.User{
			{Nickname: "宠爱动物医院", Password: "seeded", Role: "merchant", PlanType: "merchant"},
			{Nickname: "毛球美容屋", Password: "seeded", Role: "merchant", PlanType: "merchant"},
			{Nickname: "安心寄养中心", Password: "seeded", Role: "merchant", PlanType: "merchant"},
		}

		for index := range seedUsers {
			if err := tx.Create(&seedUsers[index]).Error; err != nil {
				return err
			}
		}

		providers := []model.ServiceProvider{
			{
				UserID:      seedUsers[0].ID,
				Name:        "宠爱动物医院",
				Type:        "vet_clinic",
				Description: "提供常规体检、疫苗、影像和急诊咨询。",
				Address:     "上海市徐汇区龙华中路 201 号",
				Latitude:    31.1837,
				Longitude:   121.4478,
				Phone:       "021-55550001",
				Rating:      4.8,
				ReviewCount: 138,
				IsVerified:  true,
				OpenHours:   datatypes.JSON(dto.EncodeMap(map[string]any{"daily": "09:00-21:00"})),
				Services: datatypes.JSON(dto.EncodeMapSlice([]map[string]any{
					{"name": "年度体检", "price": 199},
					{"name": "疫苗接种", "price": 128},
					{"name": "在线复诊", "price": 89},
				})),
				Tags:   datatypes.JSON(dto.EncodeStringArray([]string{"24h 咨询", "影像检查", "疫苗"})),
				Photos: datatypes.JSON(dto.EncodeStringArray([]string{})),
			},
			{
				UserID:      seedUsers[1].ID,
				Name:        "毛球美容屋",
				Type:        "grooming",
				Description: "洗护、美容、修爪和上门接送。",
				Address:     "上海市静安区愚园路 108 号",
				Latitude:    31.2228,
				Longitude:   121.4392,
				Phone:       "021-55550002",
				Rating:      4.6,
				ReviewCount: 96,
				IsVerified:  true,
				OpenHours:   datatypes.JSON(dto.EncodeMap(map[string]any{"daily": "10:00-20:00"})),
				Services: datatypes.JSON(dto.EncodeMapSlice([]map[string]any{
					{"name": "基础洗护", "price": 129},
					{"name": "造型美容", "price": 228},
					{"name": "上门接送", "price": 50},
				})),
				Tags:   datatypes.JSON(dto.EncodeStringArray([]string{"猫犬通用", "上门接送", "会员折扣"})),
				Photos: datatypes.JSON(dto.EncodeStringArray([]string{})),
			},
			{
				UserID:      seedUsers[2].ID,
				Name:        "安心寄养中心",
				Type:        "boarding",
				Description: "短住、长住和白天托管，支持实时视频查看。",
				Address:     "上海市浦东新区杨高中路 1860 号",
				Latitude:    31.2284,
				Longitude:   121.5578,
				Phone:       "021-55550003",
				Rating:      4.7,
				ReviewCount: 71,
				IsVerified:  true,
				OpenHours:   datatypes.JSON(dto.EncodeMap(map[string]any{"daily": "08:00-22:00"})),
				Services: datatypes.JSON(dto.EncodeMapSlice([]map[string]any{
					{"name": "白天托管", "price": 168},
					{"name": "单日寄养", "price": 260},
					{"name": "视频看护", "price": 39},
				})),
				Tags:   datatypes.JSON(dto.EncodeStringArray([]string{"视频监控", "每日遛弯", "长期寄养"})),
				Photos: datatypes.JSON(dto.EncodeStringArray([]string{})),
			},
		}

		return tx.Create(&providers).Error
	})
}

func (r *ServiceProviderRepository) ListProviders(ctx context.Context, typeFilter string) ([]model.ServiceProvider, error) {
	var providers []model.ServiceProvider
	query := r.db.WithContext(ctx).Model(&model.ServiceProvider{})
	if typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}
	err := query.Order("rating DESC, created_at DESC").Find(&providers).Error
	if err != nil {
		return nil, err
	}
	return providers, nil
}

func (r *ServiceProviderRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ServiceProvider, error) {
	var provider model.ServiceProvider
	err := r.db.WithContext(ctx).First(&provider, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ServiceProviderRepository) ListRatedBookings(ctx context.Context, providerID uuid.UUID) ([]model.Booking, error) {
	var bookings []model.Booking
	err := r.db.WithContext(ctx).
		Where("provider_id = ? AND rating IS NOT NULL", providerID).
		Order("updated_at DESC").
		Limit(20).
		Find(&bookings).Error
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *ServiceProviderRepository) ListProviderBookingsBetween(ctx context.Context, providerID uuid.UUID, start, end time.Time) ([]model.Booking, error) {
	var bookings []model.Booking
	err := r.db.WithContext(ctx).
		Where("provider_id = ? AND start_time >= ? AND start_time <= ? AND status IN ?", providerID, start, end, []string{"pending", "confirmed"}).
		Find(&bookings).Error
	if err != nil {
		return nil, err
	}
	return bookings, nil
}
