package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Device struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OwnerID      uuid.UUID      `gorm:"type:uuid;index;not null"`
	PetID        *uuid.UUID     `gorm:"type:uuid;index"`
	DeviceType   string         `gorm:"size:30;not null"`
	Brand        string         `gorm:"size:50"`
	Model        string         `gorm:"size:50"`
	Nickname     string         `gorm:"size:50"`
	SerialNumber string         `gorm:"size:100;uniqueIndex"`
	Status       string         `gorm:"size:20;default:offline"`
	Config       datatypes.JSON `gorm:"type:jsonb"`
	LastSeen     *time.Time
	FirmwareVer  string `gorm:"size:20"`
	BatteryLevel *int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Owner        User           `gorm:"foreignKey:OwnerID"`
	Pet          *Pet           `gorm:"foreignKey:PetID"`
}

func (d *Device) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

type DeviceDataPoint struct {
	ID       uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Time     time.Time      `gorm:"not null;index:idx_device_data_time,priority:2"`
	DeviceID uuid.UUID      `gorm:"type:uuid;not null;index:idx_device_data_time,priority:1"`
	Metric   string         `gorm:"size:50;not null"`
	Value    float64        `gorm:"not null"`
	Unit     string         `gorm:"size:20"`
	Meta     datatypes.JSON `gorm:"type:jsonb"`
}

func (d *DeviceDataPoint) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
