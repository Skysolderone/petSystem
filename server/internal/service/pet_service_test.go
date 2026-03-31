package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
)

type fakePetRepo struct {
	pets map[uuid.UUID]*model.Pet
}

func newFakePetRepo() *fakePetRepo {
	return &fakePetRepo{
		pets: map[uuid.UUID]*model.Pet{},
	}
}

func (r *fakePetRepo) Create(_ context.Context, pet *model.Pet) error {
	if pet.ID == uuid.Nil {
		pet.ID = uuid.New()
	}
	r.pets[pet.ID] = pet
	return nil
}

func (r *fakePetRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, page, pageSize int) ([]model.Pet, int64, error) {
	var pets []model.Pet
	for _, pet := range r.pets {
		if pet.OwnerID == ownerID {
			pets = append(pets, *pet)
		}
	}
	return pets, int64(len(pets)), nil
}

func (r *fakePetRepo) GetByIDAndOwner(_ context.Context, id, ownerID uuid.UUID) (*model.Pet, error) {
	pet := r.pets[id]
	if pet == nil || pet.OwnerID != ownerID {
		return nil, nil
	}
	return pet, nil
}

func (r *fakePetRepo) Update(_ context.Context, pet *model.Pet) error {
	r.pets[pet.ID] = pet
	return nil
}

func (r *fakePetRepo) Delete(_ context.Context, pet *model.Pet) error {
	delete(r.pets, pet.ID)
	return nil
}

func TestPetServiceCreateAndList(t *testing.T) {
	t.Parallel()

	repo := newFakePetRepo()
	service := NewPetService(repo)
	ownerID := uuid.New()

	createdPet, err := service.Create(context.Background(), ownerID, dto.CreatePetRequest{
		Name:      "DouDou",
		Species:   "dog",
		Breed:     "corgi",
		Allergies: []string{"beef"},
		Notes:     "friendly",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if createdPet.OwnerID != ownerID {
		t.Fatalf("Create() owner mismatch: got %s want %s", createdPet.OwnerID, ownerID)
	}

	pets, total, page, pageSize, err := service.List(context.Background(), ownerID, 0, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 1 || len(pets) != 1 {
		t.Fatalf("List() unexpected result: total=%d len=%d", total, len(pets))
	}
	if page != 1 || pageSize != 20 {
		t.Fatalf("List() pagination mismatch: page=%d pageSize=%d", page, pageSize)
	}
}
