package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
	"petverse/server/internal/pkg/apperror"
	"petverse/server/internal/pkg/pagination"
)

type BookingService struct {
	bookings  bookingRepository
	providers bookingProviderRepository
	pets      bookingPetRepository
}

type bookingRepository interface {
	Create(ctx context.Context, booking *model.Booking) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Booking, int64, error)
	GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*model.Booking, error)
	Update(ctx context.Context, booking *model.Booking) error
}

type bookingProviderRepository interface {
	EnsureDemoProviders(ctx context.Context) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ServiceProvider, error)
}

type bookingPetRepository interface {
	GetByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*model.Pet, error)
}

func NewBookingService(bookings bookingRepository, providers bookingProviderRepository, pets bookingPetRepository) *BookingService {
	return &BookingService{
		bookings:  bookings,
		providers: providers,
		pets:      pets,
	}
}

func (s *BookingService) Create(ctx context.Context, userID, providerID uuid.UUID, req dto.CreateBookingRequest) (*model.Booking, error) {
	if err := s.providers.EnsureDemoProviders(ctx); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "seed_service_providers_failed", "failed to seed service providers", err)
	}
	provider, err := s.providers.GetByID(ctx, providerID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "get_service_provider_failed", "failed to load service provider", err)
	}
	if provider == nil {
		return nil, apperror.New(http.StatusNotFound, "service_provider_not_found", "service provider not found")
	}

	petID, err := uuid.Parse(req.PetID)
	if err != nil {
		return nil, apperror.New(http.StatusBadRequest, "invalid_pet_id", "pet id is invalid")
	}
	pet, err := s.pets.GetByIDAndOwner(ctx, petID, userID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "pet_lookup_failed", "failed to load pet", err)
	}
	if pet == nil {
		return nil, apperror.New(http.StatusNotFound, "pet_not_found", "pet not found")
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, apperror.New(http.StatusBadRequest, "invalid_start_time", "start time must be RFC3339")
	}

	booking := &model.Booking{
		UserID:      userID,
		PetID:       pet.ID,
		ProviderID:  providerID,
		ServiceName: req.ServiceName,
		Status:      "confirmed",
		StartTime:   startTime,
		Price:       req.Price,
		Currency:    "CNY",
		Notes:       req.Notes,
	}

	if err := s.bookings.Create(ctx, booking); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "create_booking_failed", "failed to create booking", err)
	}
	return booking, nil
}

func (s *BookingService) List(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Booking, int64, int, int, error) {
	page, pageSize = pagination.Normalize(page, pageSize)
	bookings, total, err := s.bookings.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, apperror.Wrap(http.StatusInternalServerError, "list_bookings_failed", "failed to load bookings", err)
	}
	return bookings, total, page, pageSize, nil
}

func (s *BookingService) Get(ctx context.Context, userID, bookingID uuid.UUID) (*model.Booking, error) {
	booking, err := s.bookings.GetByIDAndUser(ctx, bookingID, userID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "get_booking_failed", "failed to load booking", err)
	}
	if booking == nil {
		return nil, apperror.New(http.StatusNotFound, "booking_not_found", "booking not found")
	}
	return booking, nil
}

func (s *BookingService) Cancel(ctx context.Context, userID, bookingID uuid.UUID, reason string) (*model.Booking, error) {
	booking, err := s.Get(ctx, userID, bookingID)
	if err != nil {
		return nil, err
	}
	booking.Status = "cancelled"
	booking.CancelReason = reason
	if err := s.bookings.Update(ctx, booking); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "cancel_booking_failed", "failed to cancel booking", err)
	}
	return booking, nil
}

func (s *BookingService) Review(ctx context.Context, userID, bookingID uuid.UUID, rating int, review string) (*model.Booking, error) {
	booking, err := s.Get(ctx, userID, bookingID)
	if err != nil {
		return nil, err
	}
	booking.Status = "completed"
	booking.Rating = &rating
	booking.Review = review
	if err := s.bookings.Update(ctx, booking); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "review_booking_failed", "failed to submit review", err)
	}
	return booking, nil
}
