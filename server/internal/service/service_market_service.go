package service

import (
	"context"
	"math"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"

	"petverse/server/internal/model"
	"petverse/server/internal/pkg/apperror"
)

type ServiceMarketService struct {
	providers serviceProviderRepository
}

type serviceProviderRepository interface {
	EnsureDemoProviders(ctx context.Context) error
	ListProviders(ctx context.Context, typeFilter string) ([]model.ServiceProvider, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.ServiceProvider, error)
	ListRatedBookings(ctx context.Context, providerID uuid.UUID) ([]model.Booking, error)
	ListProviderBookingsBetween(ctx context.Context, providerID uuid.UUID, start, end time.Time) ([]model.Booking, error)
}

func NewServiceMarketService(providers serviceProviderRepository) *ServiceMarketService {
	return &ServiceMarketService{providers: providers}
}

func (s *ServiceMarketService) List(ctx context.Context, lat, lng *float64, typeFilter string) ([]model.ServiceProvider, map[uuid.UUID]float64, error) {
	if err := s.providers.EnsureDemoProviders(ctx); err != nil {
		return nil, nil, apperror.Wrap(http.StatusInternalServerError, "seed_service_providers_failed", "failed to seed service providers", err)
	}

	providers, err := s.providers.ListProviders(ctx, typeFilter)
	if err != nil {
		return nil, nil, apperror.Wrap(http.StatusInternalServerError, "list_service_providers_failed", "failed to load service providers", err)
	}

	distances := map[uuid.UUID]float64{}
	if lat != nil && lng != nil {
		slices.SortFunc(providers, func(left, right model.ServiceProvider) int {
			leftDistance := haversineKm(*lat, *lng, left.Latitude, left.Longitude)
			rightDistance := haversineKm(*lat, *lng, right.Latitude, right.Longitude)
			distances[left.ID] = leftDistance
			distances[right.ID] = rightDistance
			switch {
			case leftDistance < rightDistance:
				return -1
			case leftDistance > rightDistance:
				return 1
			default:
				return 0
			}
		})
	}

	return providers, distances, nil
}

func (s *ServiceMarketService) Get(ctx context.Context, providerID uuid.UUID) (*model.ServiceProvider, error) {
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
	return provider, nil
}

func (s *ServiceMarketService) Reviews(ctx context.Context, providerID uuid.UUID) ([]model.Booking, error) {
	if _, err := s.Get(ctx, providerID); err != nil {
		return nil, err
	}
	reviews, err := s.providers.ListRatedBookings(ctx, providerID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "list_service_reviews_failed", "failed to load service reviews", err)
	}
	return reviews, nil
}

func (s *ServiceMarketService) Availability(ctx context.Context, providerID uuid.UUID) ([]time.Time, error) {
	if _, err := s.Get(ctx, providerID); err != nil {
		return nil, err
	}

	now := time.Now()
	bookings, err := s.providers.ListProviderBookingsBetween(ctx, providerID, now, now.Add(7*24*time.Hour))
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "list_service_availability_failed", "failed to load availability", err)
	}

	blocked := map[string]struct{}{}
	for _, booking := range bookings {
		blocked[booking.StartTime.Format(time.RFC3339)] = struct{}{}
	}

	var slots []time.Time
	hours := []int{10, 13, 16, 19}
	for day := 0; day < 5; day++ {
		base := time.Date(now.Year(), now.Month(), now.Day()+day, 0, 0, 0, 0, now.Location())
		for _, hour := range hours {
			slot := base.Add(time.Duration(hour) * time.Hour)
			if slot.Before(now.Add(90 * time.Minute)) {
				continue
			}
			if _, exists := blocked[slot.Format(time.RFC3339)]; exists {
				continue
			}
			slots = append(slots, slot)
		}
	}

	return slots, nil
}

func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371
	latDistance := degreesToRadians(lat2 - lat1)
	lngDistance := degreesToRadians(lng2 - lng1)

	a := math.Sin(latDistance/2)*math.Sin(latDistance/2) +
		math.Cos(degreesToRadians(lat1))*math.Cos(degreesToRadians(lat2))*
			math.Sin(lngDistance/2)*math.Sin(lngDistance/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
