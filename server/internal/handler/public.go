package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"feedbacker/internal/model"
	"feedbacker/internal/service"
)

const maxBodyBytes = 32 << 10

// PublicHandler 处理公开的反馈提交请求。
type PublicHandler struct {
	svc    *service.FeedbackService
	logger *slog.Logger
}

// NewPublicHandler 创建公开接口处理器。
func NewPublicHandler(svc *service.FeedbackService, logger *slog.Logger) *PublicHandler {
	return &PublicHandler{svc: svc, logger: logger}
}

// Submit 接收并保存一条反馈。
func (h *PublicHandler) Submit(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	var req model.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			respondError(c, http.StatusRequestEntityTooLarge, "payload_too_large", "请求体过大，最大支持 32KB")
			return
		}
		respondError(c, http.StatusBadRequest, "invalid_argument", "请求体必须是合法的 JSON 对象")
		return
	}

	feedback, err := h.svc.Submit(c.Request.Context(), req, model.RequestMeta{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	var validationErr *service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		respondError(c, http.StatusBadRequest, "invalid_argument", validationErr.Message)
		return
	case err != nil:
		h.logger.Error("保存反馈失败", "request_id", c.GetString("request_id"), "error", err)
		respondError(c, http.StatusInternalServerError, "internal", "服务器内部错误，请稍后重试")
		return
	}

	c.JSON(http.StatusCreated, model.SubmitResponse{
		ID:        feedback.ID.Hex(),
		CreatedAt: feedback.CreatedAt,
	})
}
