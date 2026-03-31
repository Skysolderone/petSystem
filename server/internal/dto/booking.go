package dto

import (
	"time"

	"petverse/server/internal/model"
)

type CreateBookingRequest struct {
	ProviderID  string `json:"provider_id" binding:"required"`
	PetID       string `json:"pet_id" binding:"required"`
	ServiceName string `json:"service_name" binding:"required,min=1,max=100"`
	StartTime   string `json:"start_time" binding:"required"`
	Price       int64  `json:"price"`
	Notes       string `json:"notes"`
}

type CancelBookingRequest struct {
	Reason string `json:"reason" binding:"required,min=1,max=300"`
}

type ReviewBookingRequest struct {
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
	Review string `json:"review"`
}

type BookingResponse struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	PetID        string     `json:"pet_id"`
	ProviderID   string     `json:"provider_id"`
	ServiceName  string     `json:"service_name"`
	Status       string     `json:"status"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Price        int64      `json:"price"`
	Currency     string     `json:"currency"`
	Notes        string     `json:"notes"`
	CancelReason string     `json:"cancel_reason"`
	Rating       *int       `json:"rating,omitempty"`
	Review       string     `json:"review"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func ToBookingResponse(booking *model.Booking) BookingResponse {
	return BookingResponse{
		ID:           booking.ID.String(),
		UserID:       booking.UserID.String(),
		PetID:        booking.PetID.String(),
		ProviderID:   booking.ProviderID.String(),
		ServiceName:  booking.ServiceName,
		Status:       booking.Status,
		StartTime:    booking.StartTime,
		EndTime:      booking.EndTime,
		Price:        booking.Price,
		Currency:     booking.Currency,
		Notes:        booking.Notes,
		CancelReason: booking.CancelReason,
		Rating:       booking.Rating,
		Review:       booking.Review,
		CreatedAt:    booking.CreatedAt,
		UpdatedAt:    booking.UpdatedAt,
	}
}
