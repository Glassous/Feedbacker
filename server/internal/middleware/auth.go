package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"feedbacker/internal/model"
	"feedbacker/internal/service"
)

// AdminContextKey 是当前管理员在上下文中的键名。
const AdminContextKey = "admin"

// Auth 校验 Bearer 令牌并把管理员写入上下文。
func Auth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "unauthorized",
				"message": "缺少访问令牌",
			})
			return
		}

		admin, err := authService.Authenticate(c.Request.Context(), strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "unauthorized",
				"message": "登录状态已失效，请重新登录",
			})
			return
		}

		c.Set(AdminContextKey, admin)
		c.Next()
	}
}

// RequirePasswordChanged 阻止使用一次性密码的账号访问业务接口。
func RequirePasswordChanged() gin.HandlerFunc {
	return func(c *gin.Context) {
		if admin := CurrentAdmin(c); admin != nil && admin.MustChangePassword {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "password_change_required",
				"message": "请先修改一次性密码",
			})
			return
		}
		c.Next()
	}
}

// CurrentAdmin 返回上下文中的管理员。
func CurrentAdmin(c *gin.Context) *model.Admin {
	value, ok := c.Get(AdminContextKey)
	if !ok {
		return nil
	}
	admin, _ := value.(*model.Admin)
	return admin
}
