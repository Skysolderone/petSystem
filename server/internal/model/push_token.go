package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PushToken struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `gorm:"type:uuid;index;not null"`
	Provider   string    `gorm:"size:30;not null;default:expo"`
	Token      string    `gorm:"uniqueIndex;size:255;not null"`
	Platform   string    `gorm:"size:20;not null;default:unknown"`
	IsActive   bool      `gorm:"not null;default:true"`
	LastSeenAt time.Time `gorm:"not null;default:now()"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	User       User           `gorm:"foreignKey:UserID"`
}

func (p *PushToken) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
