package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"feedbacker/internal/middleware"
	"feedbacker/internal/model"
	"feedbacker/internal/service"
)

// AuthHandler 处理管理员登录与账号相关请求。
type AuthHandler struct {
	svc    *service.AuthService
	logger *slog.Logger
}

// NewAuthHandler 创建认证接口处理器。
func NewAuthHandler(svc *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, logger: logger}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login 校验密码并返回访问令牌。
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_argument", "请求体必须是合法的 JSON 对象")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		respondError(c, http.StatusBadRequest, "invalid_argument", "请输入密码")
		return
	}

	result, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		respondError(c, http.StatusUnauthorized, "unauthorized", "用户名或密码错误")
		return
	case err != nil:
		h.logger.Error("登录失败", "request_id", c.GetString("request_id"), "error", err)
		respondError(c, http.StatusInternalServerError, "internal", "服务器内部错误，请稍后重试")
		return
	}

	c.JSON(http.StatusOK, result)
}

// Me 返回当前登录账号信息。
func (h *AuthHandler) Me(c *gin.Context) {
	admin := middleware.CurrentAdmin(c)
	if admin == nil {
		respondError(c, http.StatusUnauthorized, "unauthorized", "登录状态已失效")
		return
	}
	c.JSON(http.StatusOK, model.AdminInfo{
		Username:           admin.Username,
		MustChangePassword: admin.MustChangePassword,
	})
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// ChangePassword 修改管理员密码并返回新令牌。
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	admin := middleware.CurrentAdmin(c)
	if admin == nil {
		respondError(c, http.StatusUnauthorized, "unauthorized", "登录状态已失效")
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_argument", "请求体必须是合法的 JSON 对象")
		return
	}

	result, err := h.svc.ChangePassword(c.Request.Context(), admin.ID, req.OldPassword, req.NewPassword)

	var validationErr *service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		respondError(c, http.StatusBadRequest, "invalid_argument", validationErr.Message)
		return
	case errors.Is(err, service.ErrInvalidCredentials):
		respondError(c, http.StatusUnauthorized, "unauthorized", "当前密码不正确")
		return
	case err != nil:
		h.logger.Error("修改密码失败", "request_id", c.GetString("request_id"), "error", err)
		respondError(c, http.StatusInternalServerError, "internal", "服务器内部错误，请稍后重试")
		return
	}

	c.JSON(http.StatusOK, result)
}

// Logout 为前端提供对称的登出入口，令牌由客户端丢弃。
func (h *AuthHandler) Logout(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
