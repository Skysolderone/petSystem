package dto

import (
	"encoding/json"
	"time"

	"petverse/server/internal/model"
)

type CreatePetRequest struct {
	Name       string     `json:"name" binding:"required,min=1,max=50"`
	Species    string     `json:"species" binding:"required,min=1,max=20"`
	Breed      string     `json:"breed"`
	Gender     string     `json:"gender"`
	BirthDate  *time.Time `json:"birth_date"`
	Weight     *float64   `json:"weight"`
	AvatarURL  string     `json:"avatar_url"`
	Microchip  *string    `json:"microchip"`
	IsNeutered bool       `json:"is_neutered"`
	Allergies  []string   `json:"allergies"`
	Notes      string     `json:"notes"`
}

type UpdatePetRequest struct {
	Name       *string    `json:"name"`
	Species    *string    `json:"species"`
	Breed      *string    `json:"breed"`
	Gender     *string    `json:"gender"`
	BirthDate  *time.Time `json:"birth_date"`
	Weight     *float64   `json:"weight"`
	AvatarURL  *string    `json:"avatar_url"`
	Microchip  *string    `json:"microchip"`
	IsNeutered *bool      `json:"is_neutered"`
	Allergies  []string   `json:"allergies"`
	Notes      *string    `json:"notes"`
}

type PetResponse struct {
	ID          string     `json:"id"`
	OwnerID     string     `json:"owner_id"`
	Name        string     `json:"name"`
	Species     string     `json:"species"`
	Breed       string     `json:"breed"`
	Gender      string     `json:"gender"`
	BirthDate   *time.Time `json:"birth_date,omitempty"`
	Weight      *float64   `json:"weight,omitempty"`
	AvatarURL   string     `json:"avatar_url"`
	Microchip   *string    `json:"microchip,omitempty"`
	IsNeutered  bool       `json:"is_neutered"`
	Allergies   []string   `json:"allergies"`
	Notes       string     `json:"notes"`
	HealthScore *int       `json:"health_score,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func ToPetResponse(pet *model.Pet) PetResponse {
	return PetResponse{
		ID:          pet.ID.String(),
		OwnerID:     pet.OwnerID.String(),
		Name:        pet.Name,
		Species:     pet.Species,
		Breed:       pet.Breed,
		Gender:      pet.Gender,
		BirthDate:   pet.BirthDate,
		Weight:      pet.Weight,
		AvatarURL:   pet.AvatarURL,
		Microchip:   pet.Microchip,
		IsNeutered:  pet.IsNeutered,
		Allergies:   decodeStringArray(pet.Allergies),
		Notes:       pet.Notes,
		HealthScore: pet.HealthScore,
		CreatedAt:   pet.CreatedAt,
		UpdatedAt:   pet.UpdatedAt,
	}
}

func EncodeStringArray(values []string) []byte {
	if len(values) == 0 {
		return []byte("[]")
	}
	payload, err := json.Marshal(values)
	if err != nil {
		return []byte("[]")
	}
	return payload
}

func decodeStringArray(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}

	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return []string{}
	}
	return values
}
