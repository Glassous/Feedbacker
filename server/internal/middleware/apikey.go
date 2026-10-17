package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIKey 在配置了共享密钥时校验 X-API-Key 请求头。
func APIKey(expected string) gin.HandlerFunc {
	if expected == "" {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		provided := c.GetHeader("X-API-Key")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "unauthorized",
				"message": "API Key 无效",
			})
			return
		}
		c.Next()
	}
}
