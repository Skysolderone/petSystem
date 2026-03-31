package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"petverse/server/internal/dto"
	"petverse/server/internal/pkg/response"
	"petverse/server/internal/pkg/upload"
)

type UploadHandler struct {
	uploader upload.Store
}

func NewUploadHandler(uploader upload.Store) *UploadHandler {
	return &UploadHandler{uploader: uploader}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, badRequest("missing image file"))
		return
	}

	storedFile, err := h.uploader.SaveImage("images", fileHeader)
	if err != nil {
		response.Error(c, badRequest("invalid image file"))
		return
	}

	response.Success(c, http.StatusCreated, dto.UploadResponse{
		URL:      resolveStoredFileURL(c, storedFile),
		Path:     storedFile.RelativePath,
		Filename: storedFile.Filename,
		Size:     storedFile.Size,
		MIMEType: storedFile.MIMEType,
	}, nil)
}

func (h *UploadHandler) UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, badRequest("missing file"))
		return
	}

	storedFile, err := h.uploader.SaveFile("files", fileHeader)
	if err != nil {
		response.Error(c, badRequest("invalid file"))
		return
	}

	response.Success(c, http.StatusCreated, dto.UploadResponse{
		URL:      resolveStoredFileURL(c, storedFile),
		Path:     storedFile.RelativePath,
		Filename: storedFile.Filename,
		Size:     storedFile.Size,
		MIMEType: storedFile.MIMEType,
	}, nil)
}

func resolveStoredFileURL(c *gin.Context, storedFile *upload.StoredFile) string {
	if storedFile == nil {
		return ""
	}
	if storedFile.PublicURL != "" {
		return storedFile.PublicURL
	}
	return publicFileURL(c, storedFile.RelativePath)
}

func publicFileURL(c *gin.Context, relativePath string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if forwarded := c.GetHeader("X-Forwarded-Proto"); forwarded != "" {
		scheme = forwarded
	}
	return fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, relativePath)
}
