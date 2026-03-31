package dto

import (
	"time"

	"petverse/server/internal/model"
)

type UserResponse struct {
	ID         string     `json:"id"`
	Phone      *string    `json:"phone,omitempty"`
	Email      *string    `json:"email,omitempty"`
	Nickname   string     `json:"nickname"`
	AvatarURL  string     `json:"avatar_url"`
	Latitude   *float64   `json:"latitude,omitempty"`
	Longitude  *float64   `json:"longitude,omitempty"`
	Role       string     `json:"role"`
	PlanType   string     `json:"plan_type"`
	PlanExpiry *time.Time `json:"plan_expiry,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func ToUserResponse(user *model.User) UserResponse {
	return UserResponse{
		ID:         user.ID.String(),
		Phone:      user.Phone,
		Email:      user.Email,
		Nickname:   user.Nickname,
		AvatarURL:  user.AvatarURL,
		Latitude:   user.Latitude,
		Longitude:  user.Longitude,
		Role:       user.Role,
		PlanType:   user.PlanType,
		PlanExpiry: user.PlanExpiry,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}
