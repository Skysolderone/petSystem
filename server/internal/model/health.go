package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type HealthRecord struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PetID       uuid.UUID      `gorm:"type:uuid;index;not null"`
	Type        string         `gorm:"size:30;not null"`
	Title       string         `gorm:"size:200;not null"`
	Description string         `gorm:"type:text"`
	Data        datatypes.JSON `gorm:"type:jsonb"`
	DueDate     *time.Time
	ProviderID  *uuid.UUID     `gorm:"type:uuid"`
	Attachments datatypes.JSON `gorm:"type:jsonb"`
	RecordedAt  time.Time      `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Pet         Pet            `gorm:"foreignKey:PetID"`
}

func (h *HealthRecord) BeforeCreate(_ *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

type HealthAlert struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PetID       uuid.UUID `gorm:"type:uuid;index;not null"`
	AlertType   string    `gorm:"size:30;not null"`
	Severity    string    `gorm:"size:10;not null"`
	Title       string    `gorm:"size:200;not null"`
	Message     string    `gorm:"type:text;not null"`
	Source      string    `gorm:"size:30"`
	IsRead      bool      `gorm:"default:false"`
	IsDismissed bool      `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Pet         Pet            `gorm:"foreignKey:PetID"`
}

func (h *HealthAlert) BeforeCreate(_ *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}
