package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ServiceProvider struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	Name        string    `gorm:"size:100;not null"`
	Type        string    `gorm:"size:30;not null"`
	Description string    `gorm:"type:text"`
	Address     string    `gorm:"size:300"`
	Latitude    float64
	Longitude   float64
	Phone       string         `gorm:"size:20"`
	Photos      datatypes.JSON `gorm:"type:jsonb"`
	Rating      float64        `gorm:"default:0"`
	ReviewCount int            `gorm:"default:0"`
	IsVerified  bool           `gorm:"default:false"`
	OpenHours   datatypes.JSON `gorm:"type:jsonb"`
	Services    datatypes.JSON `gorm:"type:jsonb"`
	Tags        datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	User        User           `gorm:"foreignKey:UserID"`
}

func (s *ServiceProvider) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type Booking struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;index;not null"`
	PetID        uuid.UUID `gorm:"type:uuid;index;not null"`
	ProviderID   uuid.UUID `gorm:"type:uuid;index;not null"`
	ServiceName  string    `gorm:"size:100;not null"`
	Status       string    `gorm:"size:20;default:pending"`
	StartTime    time.Time `gorm:"not null"`
	EndTime      *time.Time
	Price        int64
	Currency     string `gorm:"size:3;default:CNY"`
	Notes        string `gorm:"type:text"`
	CancelReason string `gorm:"type:text"`
	Rating       *int
	Review       string `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt  `gorm:"index"`
	User         User            `gorm:"foreignKey:UserID"`
	Pet          Pet             `gorm:"foreignKey:PetID"`
	Provider     ServiceProvider `gorm:"foreignKey:ProviderID"`
}

func (b *Booking) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
