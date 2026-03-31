package dto

import (
	"time"

	"petverse/server/internal/model"
)

type ProductResponse struct {
	ID                string    `json:"id"`
	ProviderID        *string   `json:"provider_id,omitempty"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Category          string    `json:"category"`
	Price             int64     `json:"price"`
	Currency          string    `json:"currency"`
	Images            []string  `json:"images"`
	PetSpecies        []string  `json:"pet_species"`
	Tags              []string  `json:"tags"`
	ExternalURL       *string   `json:"external_url,omitempty"`
	Rating            float64   `json:"rating"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	RecommendedReason string    `json:"recommended_reason,omitempty"`
}

func ToProductResponse(product *model.Product) ProductResponse {
	var providerID *string
	if product.ProviderID != nil {
		value := product.ProviderID.String()
		providerID = &value
	}

	return ProductResponse{
		ID:          product.ID.String(),
		ProviderID:  providerID,
		Name:        product.Name,
		Description: product.Description,
		Category:    product.Category,
		Price:       product.Price,
		Currency:    product.Currency,
		Images:      decodeStringArray(product.Images),
		PetSpecies:  decodeStringArray(product.PetSpecies),
		Tags:        decodeStringArray(product.Tags),
		ExternalURL: product.ExternalURL,
		Rating:      product.Rating,
		IsActive:    product.IsActive,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}
