package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
)

type ShopRepository struct {
	db *gorm.DB
}

func NewShopRepository(db *gorm.DB) *ShopRepository {
	return &ShopRepository{db: db}
}

func (r *ShopRepository) EnsureDemoProducts(ctx context.Context) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Product{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	products := []model.Product{
		{
			Name:        "低敏鸡肉主粮",
			Description: "适合狗狗日常营养管理，强调消化友好和均衡蛋白。",
			Category:    "food",
			Price:       12900,
			Currency:    "CNY",
			Images:      dto.EncodeStringArray([]string{}),
			PetSpecies:  dto.EncodeStringArray([]string{"dog"}),
			Tags:        dto.EncodeStringArray([]string{"低敏", "日常主粮"}),
			Rating:      4.8,
			IsActive:    true,
		},
		{
			Name:        "猫咪泌尿健康湿粮",
			Description: "针对饮水偏少场景设计的猫咪湿粮，适合补水辅助。",
			Category:    "health",
			Price:       15900,
			Currency:    "CNY",
			Images:      dto.EncodeStringArray([]string{}),
			PetSpecies:  dto.EncodeStringArray([]string{"cat"}),
			Tags:        dto.EncodeStringArray([]string{"补水", "泌尿护理"}),
			Rating:      4.7,
			IsActive:    true,
		},
		{
			Name:        "耐咬发声玩具球",
			Description: "用于消耗精力和互动训练，适合新手家庭做奖励。",
			Category:    "toys",
			Price:       4900,
			Currency:    "CNY",
			Images:      dto.EncodeStringArray([]string{}),
			PetSpecies:  dto.EncodeStringArray([]string{"dog", "cat"}),
			Tags:        dto.EncodeStringArray([]string{"互动", "训练奖励"}),
			Rating:      4.6,
			IsActive:    true,
		},
		{
			Name:        "宠物梳毛护理套装",
			Description: "适合长毛宠物的基础美容护理，减少打结和浮毛。",
			Category:    "grooming",
			Price:       8900,
			Currency:    "CNY",
			Images:      dto.EncodeStringArray([]string{}),
			PetSpecies:  dto.EncodeStringArray([]string{"dog", "cat"}),
			Tags:        dto.EncodeStringArray([]string{"长毛护理", "美容"}),
			Rating:      4.5,
			IsActive:    true,
		},
		{
			Name:        "关节营养补充剂",
			Description: "适合中大型犬日常活动支持，也可用于恢复期营养补充。",
			Category:    "health",
			Price:       19900,
			Currency:    "CNY",
			Images:      dto.EncodeStringArray([]string{}),
			PetSpecies:  dto.EncodeStringArray([]string{"dog"}),
			Tags:        dto.EncodeStringArray([]string{"关节", "营养补充"}),
			Rating:      4.9,
			IsActive:    true,
		},
	}

	return r.db.WithContext(ctx).Create(&products).Error
}

func (r *ShopRepository) ListProducts(ctx context.Context, category, query string) ([]model.Product, error) {
	var products []model.Product
	db := r.db.WithContext(ctx).Model(&model.Product{}).Where("is_active = ?", true)
	if category != "" {
		db = db.Where("category = ?", category)
	}
	if query != "" {
		pattern := "%" + strings.ToLower(query) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", pattern, pattern)
	}
	err := db.Order("rating DESC, created_at DESC").Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ShopRepository) GetByID(ctx context.Context, id string) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).First(&product, "id = ? AND is_active = ?", id, true).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func DecodeProductSpecies(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return []string{}
	}
	return values
}
