package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Product struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProviderID  *uuid.UUID `gorm:"type:uuid;index"`
	Name        string     `gorm:"size:200;not null"`
	Description string     `gorm:"type:text"`
	Category    string     `gorm:"size:30"`
	Price       int64
	Currency    string         `gorm:"size:3;default:CNY"`
	Images      datatypes.JSON `gorm:"type:jsonb"`
	PetSpecies  datatypes.JSON `gorm:"type:jsonb"`
	Tags        datatypes.JSON `gorm:"type:jsonb"`
	ExternalURL *string        `gorm:"size:500"`
	Rating      float64        `gorm:"default:0"`
	IsActive    bool           `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (p *Product) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
