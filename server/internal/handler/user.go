package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"petverse/server/internal/dto"
	"petverse/server/internal/middleware"
	"petverse/server/internal/pkg/response"
	"petverse/server/internal/pkg/upload"
	"petverse/server/internal/service"
)

type UserHandler struct {
	users    *service.UserService
	uploader upload.Store
}

func NewUserHandler(users *service.UserService, uploader upload.Store) *UserHandler {
	return &UserHandler{
		users:    users,
		uploader: uploader,
	}
}

func (h *UserHandler) Me(c *gin.Context) {
	user, err := h.users.GetMe(c.Request.Context(), middleware.MustUserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserResponse(user), nil)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid user payload"))
		return
	}

	user, err := h.users.UpdateMe(c.Request.Context(), middleware.MustUserID(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToUserResponse(user), nil)
}

func (h *UserHandler) UpdateLocation(c *gin.Context) {
	var req dto.UpdateUserLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid location payload"))
		return
	}

	user, err := h.users.UpdateLocation(c.Request.Context(), middleware.MustUserID(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToUserResponse(user), nil)
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	if h.uploader == nil {
		response.Error(c, badRequest("upload service is unavailable"))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, badRequest("missing avatar file"))
		return
	}

	storedFile, err := h.uploader.SaveImage("users", fileHeader)
	if err != nil {
		response.Error(c, badRequest("invalid avatar file"))
		return
	}

	user, err := h.users.UpdateAvatar(c.Request.Context(), middleware.MustUserID(c), resolveStoredFileURL(c, storedFile))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToUserResponse(user), nil)
}
