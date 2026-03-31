package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/middleware"
	"petverse/server/internal/pkg/response"
	"petverse/server/internal/service"
)

type ShopHandler struct {
	shop *service.ShopService
}

func NewShopHandler(shop *service.ShopService) *ShopHandler {
	return &ShopHandler{shop: shop}
}

func (h *ShopHandler) ListProducts(c *gin.Context) {
	products, err := h.shop.List(c.Request.Context(), c.Query("category"), c.Query("q"))
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.ProductResponse, 0, len(products))
	for _, product := range products {
		productCopy := product
		items = append(items, dto.ToProductResponse(&productCopy))
	}
	response.Success(c, http.StatusOK, items, nil)
}

func (h *ShopHandler) GetProduct(c *gin.Context) {
	product, err := h.shop.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToProductResponse(product), nil)
}

func (h *ShopHandler) Recommendations(c *gin.Context) {
	petID, err := uuid.Parse(c.Param("petId"))
	if err != nil {
		response.Error(c, badRequest("invalid pet id"))
		return
	}

	products, reasons, err := h.shop.Recommendations(c.Request.Context(), middleware.MustUserID(c), petID)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.ProductResponse, 0, len(products))
	for _, product := range products {
		productCopy := product
		item := dto.ToProductResponse(&productCopy)
		item.RecommendedReason = reasons[product.ID]
		items = append(items, item)
	}
	response.Success(c, http.StatusOK, items, nil)
}
