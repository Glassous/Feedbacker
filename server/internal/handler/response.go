// Package handler 实现 HTTP 接口层。
package handler

import "github.com/gin-gonic/gin"

// respondError 输出统一的错误结构。
func respondError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"code":    code,
		"message": message,
	})
}
