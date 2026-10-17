package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery 捕获 panic，记录堆栈并返回统一错误结构。
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("请求处理发生 panic",
					"request_id", c.GetString(RequestIDKey),
					"path", c.Request.URL.Path,
					"error", recovered,
					"stack", string(debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    "internal",
					"message": "服务器内部错误",
				})
			}
		}()
		c.Next()
	}
}
