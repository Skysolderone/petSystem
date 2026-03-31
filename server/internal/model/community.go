package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Post struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AuthorID     uuid.UUID      `gorm:"type:uuid;index;not null"`
	PetID        *uuid.UUID     `gorm:"type:uuid;index"`
	Type         string         `gorm:"size:20;default:post"`
	Title        string         `gorm:"size:200"`
	Content      string         `gorm:"type:text;not null"`
	Images       datatypes.JSON `gorm:"type:jsonb"`
	Tags         datatypes.JSON `gorm:"type:jsonb"`
	Latitude     *float64
	Longitude    *float64
	LikeCount    int  `gorm:"default:0"`
	CommentCount int  `gorm:"default:0"`
	IsPublished  bool `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Author       User           `gorm:"foreignKey:AuthorID"`
}

func (p *Post) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type Comment struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PostID    uuid.UUID  `gorm:"type:uuid;index;not null"`
	AuthorID  uuid.UUID  `gorm:"type:uuid;index;not null"`
	ParentID  *uuid.UUID `gorm:"type:uuid;index"`
	Content   string     `gorm:"type:text;not null"`
	LikeCount int        `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Author    User           `gorm:"foreignKey:AuthorID"`
}

func (c *Comment) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type PostLike struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	PostID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
}
