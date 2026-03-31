package service

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"github.com/google/uuid"

	"petverse/server/internal/model"
	"petverse/server/internal/pkg/apperror"
)

type ShopService struct {
	products shopRepository
	pets     shopPetRepository
}

type shopRepository interface {
	EnsureDemoProducts(ctx context.Context) error
	ListProducts(ctx context.Context, category, query string) ([]model.Product, error)
	GetByID(ctx context.Context, id string) (*model.Product, error)
}

type shopPetRepository interface {
	GetByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*model.Pet, error)
}

func NewShopService(products shopRepository, pets shopPetRepository) *ShopService {
	return &ShopService{
		products: products,
		pets:     pets,
	}
}

func (s *ShopService) List(ctx context.Context, category, query string) ([]model.Product, error) {
	if err := s.products.EnsureDemoProducts(ctx); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "seed_products_failed", "failed to seed products", err)
	}
	items, err := s.products.ListProducts(ctx, category, query)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "list_products_failed", "failed to load products", err)
	}
	return items, nil
}

func (s *ShopService) Get(ctx context.Context, productID string) (*model.Product, error) {
	if err := s.products.EnsureDemoProducts(ctx); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "seed_products_failed", "failed to seed products", err)
	}
	item, err := s.products.GetByID(ctx, productID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "get_product_failed", "failed to load product", err)
	}
	if item == nil {
		return nil, apperror.New(http.StatusNotFound, "product_not_found", "product not found")
	}
	return item, nil
}

func (s *ShopService) Recommendations(ctx context.Context, userID, petID uuid.UUID) ([]model.Product, map[uuid.UUID]string, error) {
	if err := s.products.EnsureDemoProducts(ctx); err != nil {
		return nil, nil, apperror.Wrap(http.StatusInternalServerError, "seed_products_failed", "failed to seed products", err)
	}

	pet, err := s.pets.GetByIDAndOwner(ctx, petID, userID)
	if err != nil {
		return nil, nil, apperror.Wrap(http.StatusInternalServerError, "pet_lookup_failed", "failed to load pet", err)
	}
	if pet == nil {
		return nil, nil, apperror.New(http.StatusNotFound, "pet_not_found", "pet not found")
	}

	items, err := s.products.ListProducts(ctx, "", "")
	if err != nil {
		return nil, nil, apperror.Wrap(http.StatusInternalServerError, "list_products_failed", "failed to load products", err)
	}

	filtered := make([]model.Product, 0, len(items))
	reasons := map[uuid.UUID]string{}
	for _, item := range items {
		species := decodeProductSpecies(item.PetSpecies)
		if len(species) > 0 && !slices.Contains(species, pet.Species) && !slices.Contains(species, "all") {
			continue
		}
		filtered = append(filtered, item)
		reasons[item.ID] = recommendationReason(pet, item)
	}

	slices.SortFunc(filtered, func(left, right model.Product) int {
		if left.Rating > right.Rating {
			return -1
		}
		if left.Rating < right.Rating {
			return 1
		}
		return 0
	})

	if len(filtered) > 4 {
		filtered = filtered[:4]
	}
	return filtered, reasons, nil
}

func recommendationReason(pet *model.Pet, product model.Product) string {
	switch product.Category {
	case "food":
		return "结合宠物档案，优先推荐适合当前物种的日常主粮。"
	case "health":
		return "更适合用于健康管理或恢复期的日常补充。"
	case "toys":
		return "适合做互动训练和精力消耗，能提高训练计划执行度。"
	case "grooming":
		return "适合做居家护理，减少去店频率。"
	default:
		return strings.TrimSpace(pet.Name + " 当前可以从这类用品开始补齐日常配置。")
	}
}

func decodeProductSpecies(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}

	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return []string{}
	}
	return values
}
