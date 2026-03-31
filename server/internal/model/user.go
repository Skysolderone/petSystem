package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Phone        *string   `gorm:"uniqueIndex;size:20"`
	Email        *string   `gorm:"uniqueIndex;size:255"`
	Password     string    `gorm:"size:255;not null"`
	Nickname     string    `gorm:"size:50;not null"`
	AvatarURL    string    `gorm:"size:500"`
	Role         string    `gorm:"size:20;default:user"`
	WechatOpenID *string   `gorm:"uniqueIndex;size:100"`
	AppleID      *string   `gorm:"uniqueIndex;size:100"`
	GoogleID     *string   `gorm:"uniqueIndex;size:100"`
	Latitude     *float64
	Longitude    *float64
	PlanType     string `gorm:"size:20;default:free"`
	PlanExpiry   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Pets         []Pet          `gorm:"foreignKey:OwnerID"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
