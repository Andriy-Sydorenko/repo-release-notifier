package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" {
			c.Next()
			return
		}

		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "missing API key"})
			return
		}
		if key != apiKey {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{Error: "invalid API key"})
			return
		}

		c.Next()
	}
}
