// Package service 承载业务规则与领域校验。
package service

import (
	"context"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/v2/bson"

	"feedbacker/internal/model"
	"feedbacker/internal/repository"
)

const (
	maxTitleLen     = 200
	maxContentLen   = 5000
	maxContactLen   = 200
	maxProjectLen   = 64
	maxPageURLLen   = 2048
	maxUserAgent    = 512
	maxQueryLen     = 64
	defaultPageSize = 20
	maxPageSize     = 100
)

// ValidationError 表示请求参数不符合业务规则。
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

// FeedbackService 实现反馈提交与后台查询逻辑。
type FeedbackService struct {
	repo *repository.FeedbackRepository
}

// NewFeedbackService 创建反馈服务。
func NewFeedbackService(repo *repository.FeedbackRepository) *FeedbackService {
	return &FeedbackService{repo: repo}
}

// ListParams 是后台列表查询入参。
type ListParams struct {
	Page     int
	PageSize int
	Query    string
	Status   string
	Project  string
}

// ListOutput 是规范化后的列表结果。
type ListOutput struct {
	Items    []model.Feedback
	Total    int64
	Counts   model.Counts
	Page     int
	PageSize int
}

// Submit 校验并写入一条反馈。
func (s *FeedbackService) Submit(ctx context.Context, req model.SubmitRequest, meta model.RequestMeta) (*model.Feedback, error) {
	title := sanitize(req.Title)
	content := sanitize(req.Content)
	contact := sanitize(req.Contact)
	project := sanitize(req.Project)
	pageURL := sanitize(req.PageURL)

	if utf8.RuneCountInString(title) > maxTitleLen {
		return nil, &ValidationError{Message: "标题不能超过 200 个字符"}
	}
	if utf8.RuneCountInString(content) > maxContentLen {
		return nil, &ValidationError{Message: "正文不能超过 5000 个字符"}
	}
	if utf8.RuneCountInString(contact) > maxContactLen {
		return nil, &ValidationError{Message: "联系方式不能超过 200 个字符"}
	}
	if utf8.RuneCountInString(project) > maxProjectLen {
		return nil, &ValidationError{Message: "来源标识不能超过 64 个字符"}
	}
	if utf8.RuneCountInString(pageURL) > maxPageURLLen {
		return nil, &ValidationError{Message: "页面地址不能超过 2048 个字符"}
	}
	if pageURL != "" && !isHTTPURL(pageURL) {
		return nil, &ValidationError{Message: "页面地址必须是合法的 http(s) 链接"}
	}
	if title == "" && content == "" && contact == "" {
		return nil, &ValidationError{Message: "请至少填写标题、正文或联系方式中的一项"}
	}

	now := time.Now().UTC()
	feedback := &model.Feedback{
		Title:     title,
		Content:   content,
		Contact:   contact,
		Project:   project,
		PageURL:   pageURL,
		UserAgent: truncate(sanitize(meta.UserAgent), maxUserAgent),
		IP:        truncate(sanitize(meta.IP), 64),
		Status:    model.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Insert(ctx, feedback); err != nil {
		return nil, err
	}
	return feedback, nil
}

// List 返回过滤后的分页结果。
func (s *FeedbackService) List(ctx context.Context, params ListParams) (*ListOutput, error) {
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	status := strings.TrimSpace(params.Status)
	if status != "" && !model.Status(status).Valid() {
		return nil, &ValidationError{Message: "状态只能是 open 或 resolved"}
	}

	query := truncate(strings.TrimSpace(params.Query), maxQueryLen)
	project := truncate(strings.TrimSpace(params.Project), maxProjectLen)

	result, err := s.repo.List(ctx, repository.ListParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
		Query:    query,
		Status:   status,
		Project:  project,
	})
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		Items:    result.Items,
		Total:    result.Total,
		Counts:   result.Counts,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 按 ID 查询一条反馈。
func (s *FeedbackService) Get(ctx context.Context, rawID string) (*model.Feedback, error) {
	id, err := parseObjectID(rawID)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, id)
}

// UpdateStatus 修改反馈处理状态。
func (s *FeedbackService) UpdateStatus(ctx context.Context, rawID, rawStatus string) (*model.Feedback, error) {
	id, err := parseObjectID(rawID)
	if err != nil {
		return nil, err
	}
	status := model.Status(strings.TrimSpace(rawStatus))
	if !status.Valid() {
		return nil, &ValidationError{Message: "状态只能是 open 或 resolved"}
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

// Delete 删除一条反馈。
func (s *FeedbackService) Delete(ctx context.Context, rawID string) error {
	id, err := parseObjectID(rawID)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// Projects 返回来源项目列表，用于后台筛选。
func (s *FeedbackService) Projects(ctx context.Context) ([]string, error) {
	return s.repo.DistinctProjects(ctx)
}

func parseObjectID(raw string) (bson.ObjectID, error) {
	id, err := bson.ObjectIDFromHex(strings.TrimSpace(raw))
	if err != nil {
		return bson.NilObjectID, &ValidationError{Message: "无效的反馈 ID"}
	}
	return id, nil
}

func isHTTPURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return parsed.Host != ""
}

// sanitize 去除首尾空白并向内剔除控制字符，避免污染存储与展示。
func sanitize(raw string) string {
	cleaned := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, raw)
	return strings.TrimSpace(cleaned)
}

func truncate(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	runes := []rune(s)
	return string(runes[:limit])
}

// ErrNotFound 表示目标记录不存在。
var ErrNotFound = repository.ErrNotFound
