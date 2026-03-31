package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TrainingPlan struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PetID       uuid.UUID      `gorm:"type:uuid;index;not null"`
	Title       string         `gorm:"size:100;not null"`
	Description string         `gorm:"type:text"`
	Difficulty  string         `gorm:"size:20"`
	Category    string         `gorm:"size:30"`
	Steps       datatypes.JSON `gorm:"type:jsonb"`
	AIGenerated bool           `gorm:"default:false"`
	Progress    int            `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Pet         Pet            `gorm:"foreignKey:PetID"`
}

func (t *TrainingPlan) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
