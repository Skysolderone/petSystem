package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/middleware"
	"petverse/server/internal/pkg/response"
	"petverse/server/internal/pkg/upload"
	"petverse/server/internal/service"
)

type PetHandler struct {
	pets     *service.PetService
	uploader upload.Store
}

func NewPetHandler(pets *service.PetService, uploader upload.Store) *PetHandler {
	return &PetHandler{
		pets:     pets,
		uploader: uploader,
	}
}

func (h *PetHandler) Create(c *gin.Context) {
	var req dto.CreatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid pet payload"))
		return
	}

	pet, err := h.pets.Create(c.Request.Context(), middleware.MustUserID(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToPetResponse(pet), nil)
}

func (h *PetHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	pets, total, normalizedPage, normalizedPageSize, err := h.pets.List(c.Request.Context(), middleware.MustUserID(c), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	responses := make([]dto.PetResponse, 0, len(pets))
	for _, pet := range pets {
		petCopy := pet
		responses = append(responses, dto.ToPetResponse(&petCopy))
	}

	response.Success(c, http.StatusOK, responses, gin.H{
		"page":      normalizedPage,
		"page_size": normalizedPageSize,
		"total":     total,
	})
}

func (h *PetHandler) Get(c *gin.Context) {
	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid pet id"))
		return
	}

	pet, err := h.pets.Get(c.Request.Context(), middleware.MustUserID(c), petID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPetResponse(pet), nil)
}

func (h *PetHandler) Update(c *gin.Context) {
	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid pet id"))
		return
	}

	var req dto.UpdatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid pet payload"))
		return
	}

	pet, err := h.pets.Update(c.Request.Context(), middleware.MustUserID(c), petID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPetResponse(pet), nil)
}

func (h *PetHandler) Delete(c *gin.Context) {
	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid pet id"))
		return
	}

	if err := h.pets.Delete(c.Request.Context(), middleware.MustUserID(c), petID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"deleted": true}, nil)
}

func (h *PetHandler) UploadAvatar(c *gin.Context) {
	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid pet id"))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, badRequest("missing avatar file"))
		return
	}

	storedFile, err := h.uploader.SaveImage("pets", fileHeader)
	if err != nil {
		response.Error(c, badRequest("invalid avatar file"))
		return
	}

	pet, err := h.pets.UpdateAvatar(c.Request.Context(), middleware.MustUserID(c), petID, resolveStoredFileURL(c, storedFile))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToPetResponse(pet), nil)
}
