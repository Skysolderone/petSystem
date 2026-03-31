package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"petverse/server/internal/pkg/apperror"
	petjwt "petverse/server/internal/pkg/jwt"
)

const UserIDKey = "user_id"

func Auth(manager *petjwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			rawToken = c.Query("access_token")
		}
		if rawToken == "" {
			c.Error(apperror.New(http.StatusUnauthorized, "missing_token", "missing or invalid bearer token"))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "missing or invalid bearer token",
				"data": gin.H{
					"error_code": "missing_token",
				},
			})
			return
		}

		claims, err := manager.ParseToken(rawToken)
		if err != nil || claims.TokenType != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "invalid access token",
				"data": gin.H{
					"error_code": "invalid_token",
				},
			})
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "invalid access token",
				"data": gin.H{
					"error_code": "invalid_token",
				},
			})
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

func bearerToken(authHeader string) (string, bool) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	return parts[1], true
}

func MustUserID(c *gin.Context) uuid.UUID {
	if value, ok := c.Get(UserIDKey); ok {
		if userID, ok := value.(uuid.UUID); ok {
			return userID
		}
	}
	return uuid.Nil
}
