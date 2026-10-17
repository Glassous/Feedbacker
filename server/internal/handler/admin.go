package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"feedbacker/internal/model"
	"feedbacker/internal/service"
)

// AdminHandler 处理后台反馈管理请求。
type AdminHandler struct {
	svc    *service.FeedbackService
	logger *slog.Logger
}

// NewAdminHandler 创建后台接口处理器。
func NewAdminHandler(svc *service.FeedbackService, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{svc: svc, logger: logger}
}

type listResponse struct {
	Items    []model.FeedbackDTO `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
	Counts   model.Counts        `json:"counts"`
}

// List 返回反馈列表。
func (h *AdminHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	output, err := h.svc.List(c.Request.Context(), service.ListParams{
		Page:     page,
		PageSize: pageSize,
		Query:    c.Query("q"),
		Status:   c.Query("status"),
		Project:  c.Query("project"),
	})

	var validationErr *service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		respondError(c, http.StatusBadRequest, "invalid_argument", validationErr.Message)
		return
	case err != nil:
		h.logger.Error("查询反馈列表失败", "request_id", c.GetString("request_id"), "error", err)
		respondError(c, http.StatusInternalServerError, "internal", "服务器内部错误，请稍后重试")
		return
	}

	items := make([]model.FeedbackDTO, 0, len(output.Items))
	for i := range output.Items {
		items = append(items, output.Items[i].DTO())
	}

	c.JSON(http.StatusOK, listResponse{
		Items:    items,
		Total:    output.Total,
		Page:     output.Page,
		PageSize: output.PageSize,
		Counts:   output.Counts,
	})
}

// Get 返回单条反馈详情。
func (h *AdminHandler) Get(c *gin.Context) {
	feedback, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if handleFeedbackError(c, h.logger, err, "查询反馈详情失败") {
		return
	}
	c.JSON(http.StatusOK, feedback.DTO())
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

// UpdateStatus 修改反馈处理状态。
func (h *AdminHandler) UpdateStatus(c *gin.Context) {
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_argument", "请求体必须是合法的 JSON 对象")
		return
	}

	feedback, err := h.svc.UpdateStatus(c.Request.Context(), c.Param("id"), req.Status)
	if handleFeedbackError(c, h.logger, err, "更新反馈状态失败") {
		return
	}
	c.JSON(http.StatusOK, feedback.DTO())
}

// Delete 删除一条反馈。
func (h *AdminHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		handleFeedbackError(c, h.logger, err, "删除反馈失败")
		return
	}
	c.Status(http.StatusNoContent)
}

// Projects 返回来源项目列表。
func (h *AdminHandler) Projects(c *gin.Context) {
	projects, err := h.svc.Projects(c.Request.Context())
	if err != nil {
		h.logger.Error("查询来源项目失败", "request_id", c.GetString("request_id"), "error", err)
		respondError(c, http.StatusInternalServerError, "internal", "服务器内部错误，请稍后重试")
		return
	}
	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

// handleFeedbackError 将业务错误映射为 HTTP 响应，返回 true 表示已处理。
func handleFeedbackError(c *gin.Context, logger *slog.Logger, err error, action string) bool {
	if err == nil {
		return false
	}

	var validationErr *service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		respondError(c, http.StatusBadRequest, "invalid_argument", validationErr.Message)
	case errors.Is(err, service.ErrNotFound):
		respondError(c, http.StatusNotFound, "not_found", "反馈不存在或已被删除")
	default:
		logger.Error(action, "request_id", c.GetString("request_id"), "error", err)
		respondError(c, http.StatusInternalServerError, "internal", "服务器内部错误，请稍后重试")
	}
	return true
}
