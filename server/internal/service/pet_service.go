package service

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
	"petverse/server/internal/pkg/apperror"
	"petverse/server/internal/pkg/pagination"
)

type PetService struct {
	pets petRepository
}

type petRepository interface {
	Create(ctx context.Context, pet *model.Pet) error
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]model.Pet, int64, error)
	GetByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*model.Pet, error)
	Update(ctx context.Context, pet *model.Pet) error
	Delete(ctx context.Context, pet *model.Pet) error
}

func NewPetService(pets petRepository) *PetService {
	return &PetService{pets: pets}
}

func (s *PetService) Create(ctx context.Context, ownerID uuid.UUID, req dto.CreatePetRequest) (*model.Pet, error) {
	pet := &model.Pet{
		OwnerID:    ownerID,
		Name:       req.Name,
		Species:    req.Species,
		Breed:      req.Breed,
		Gender:     req.Gender,
		BirthDate:  req.BirthDate,
		Weight:     req.Weight,
		AvatarURL:  req.AvatarURL,
		Microchip:  req.Microchip,
		IsNeutered: req.IsNeutered,
		Allergies:  datatypes.JSON(dto.EncodeStringArray(req.Allergies)),
		Notes:      req.Notes,
	}

	if err := s.pets.Create(ctx, pet); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "create_pet_failed", "failed to create pet", err)
	}

	return pet, nil
}

func (s *PetService) List(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]model.Pet, int64, int, int, error) {
	page, pageSize = pagination.Normalize(page, pageSize)

	pets, total, err := s.pets.ListByOwner(ctx, ownerID, page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, apperror.Wrap(http.StatusInternalServerError, "list_pets_failed", "failed to load pets", err)
	}

	return pets, total, page, pageSize, nil
}

func (s *PetService) Get(ctx context.Context, ownerID, petID uuid.UUID) (*model.Pet, error) {
	pet, err := s.pets.GetByIDAndOwner(ctx, petID, ownerID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "get_pet_failed", "failed to load pet", err)
	}
	if pet == nil {
		return nil, apperror.New(http.StatusNotFound, "pet_not_found", "pet not found")
	}
	return pet, nil
}

func (s *PetService) Update(ctx context.Context, ownerID, petID uuid.UUID, req dto.UpdatePetRequest) (*model.Pet, error) {
	pet, err := s.Get(ctx, ownerID, petID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		pet.Name = *req.Name
	}
	if req.Species != nil {
		pet.Species = *req.Species
	}
	if req.Breed != nil {
		pet.Breed = *req.Breed
	}
	if req.Gender != nil {
		pet.Gender = *req.Gender
	}
	if req.BirthDate != nil {
		pet.BirthDate = req.BirthDate
	}
	if req.Weight != nil {
		pet.Weight = req.Weight
	}
	if req.AvatarURL != nil {
		pet.AvatarURL = *req.AvatarURL
	}
	if req.Microchip != nil {
		pet.Microchip = req.Microchip
	}
	if req.IsNeutered != nil {
		pet.IsNeutered = *req.IsNeutered
	}
	if req.Allergies != nil {
		pet.Allergies = datatypes.JSON(dto.EncodeStringArray(req.Allergies))
	}
	if req.Notes != nil {
		pet.Notes = *req.Notes
	}

	if err := s.pets.Update(ctx, pet); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "update_pet_failed", "failed to update pet", err)
	}

	return pet, nil
}

func (s *PetService) Delete(ctx context.Context, ownerID, petID uuid.UUID) error {
	pet, err := s.Get(ctx, ownerID, petID)
	if err != nil {
		return err
	}

	if err := s.pets.Delete(ctx, pet); err != nil {
		return apperror.Wrap(http.StatusInternalServerError, "delete_pet_failed", "failed to delete pet", err)
	}

	return nil
}

func (s *PetService) UpdateAvatar(ctx context.Context, ownerID, petID uuid.UUID, avatarURL string) (*model.Pet, error) {
	return s.Update(ctx, ownerID, petID, dto.UpdatePetRequest{
		AvatarURL: &avatarURL,
	})
}
