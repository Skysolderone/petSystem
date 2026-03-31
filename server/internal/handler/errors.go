package handler

import (
	"net/http"

	"petverse/server/internal/pkg/apperror"
)

func badRequest(message string) error {
	return apperror.New(http.StatusBadRequest, "bad_request", message)
}
