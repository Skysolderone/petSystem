package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Pet struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OwnerID     uuid.UUID `gorm:"type:uuid;index;not null"`
	Name        string    `gorm:"size:50;not null"`
	Species     string    `gorm:"size:20;not null"`
	Breed       string    `gorm:"size:50"`
	Gender      string    `gorm:"size:10"`
	BirthDate   *time.Time
	Weight      *float64
	AvatarURL   string         `gorm:"size:500"`
	Microchip   *string        `gorm:"size:50"`
	IsNeutered  bool           `gorm:"default:false"`
	Allergies   datatypes.JSON `gorm:"type:jsonb"`
	Notes       string         `gorm:"type:text"`
	HealthScore *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Owner       User           `gorm:"foreignKey:OwnerID"`
}

func (p *Pet) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
