package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"petverse/server/internal/pkg/apperror"
)

type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

func Success(c *gin.Context, status int, data any, meta any) {
	c.JSON(status, Envelope{
		Code:    status,
		Message: "success",
		Data:    data,
		Meta:    meta,
	})
}

func Error(c *gin.Context, err error) {
	if appErr, ok := apperror.As(err); ok {
		c.JSON(appErr.HTTPStatus, Envelope{
			Code:    appErr.HTTPStatus,
			Message: appErr.Message,
			Data: gin.H{
				"error_code": appErr.Code,
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, Envelope{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
	})
}
