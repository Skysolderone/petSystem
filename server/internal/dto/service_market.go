package dto

import (
	"encoding/json"
	"time"

	"petverse/server/internal/model"
)

type ServiceProviderResponse struct {
	ID          string           `json:"id"`
	UserID      string           `json:"user_id"`
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	Description string           `json:"description"`
	Address     string           `json:"address"`
	Latitude    float64          `json:"latitude"`
	Longitude   float64          `json:"longitude"`
	Phone       string           `json:"phone"`
	Photos      []string         `json:"photos"`
	Rating      float64          `json:"rating"`
	ReviewCount int              `json:"review_count"`
	IsVerified  bool             `json:"is_verified"`
	OpenHours   map[string]any   `json:"open_hours"`
	Services    []map[string]any `json:"services"`
	Tags        []string         `json:"tags"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DistanceKm  float64          `json:"distance_km,omitempty"`
}

type ServiceAvailabilityResponse struct {
	ProviderID string   `json:"provider_id"`
	Slots      []string `json:"slots"`
}

func ToServiceProviderResponse(provider *model.ServiceProvider) ServiceProviderResponse {
	return ServiceProviderResponse{
		ID:          provider.ID.String(),
		UserID:      provider.UserID.String(),
		Name:        provider.Name,
		Type:        provider.Type,
		Description: provider.Description,
		Address:     provider.Address,
		Latitude:    provider.Latitude,
		Longitude:   provider.Longitude,
		Phone:       provider.Phone,
		Photos:      decodeStringArray(provider.Photos),
		Rating:      provider.Rating,
		ReviewCount: provider.ReviewCount,
		IsVerified:  provider.IsVerified,
		OpenHours:   decodeMap(provider.OpenHours),
		Services:    decodeMapSlice(provider.Services),
		Tags:        decodeStringArray(provider.Tags),
		CreatedAt:   provider.CreatedAt,
		UpdatedAt:   provider.UpdatedAt,
	}
}

func decodeMapSlice(raw []byte) []map[string]any {
	if len(raw) == 0 {
		return []map[string]any{}
	}
	var result []map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return []map[string]any{}
	}
	return result
}

func EncodeMapSlice(payload []map[string]any) []byte {
	if len(payload) == 0 {
		return []byte("[]")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return []byte("[]")
	}
	return raw
}
