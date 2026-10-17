package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Status 表示一条反馈的处理状态。
type Status string

const (
	// StatusOpen 表示尚未处理。
	StatusOpen Status = "open"
	// StatusResolved 表示已处理。
	StatusResolved Status = "resolved"
)

// Valid 判断状态是否为受支持的取值。
func (s Status) Valid() bool {
	return s == StatusOpen || s == StatusResolved
}

// Feedback 是 feedback 集合的文档结构。
type Feedback struct {
	ID         bson.ObjectID `bson:"_id,omitempty"`
	Title      string        `bson:"title"`
	Content    string        `bson:"content"`
	Contact    string        `bson:"contact"`
	Project    string        `bson:"project"`
	PageURL    string        `bson:"pageUrl"`
	UserAgent  string        `bson:"userAgent"`
	IP         string        `bson:"ip"`
	Status     Status        `bson:"status"`
	CreatedAt  time.Time     `bson:"createdAt"`
	UpdatedAt  time.Time     `bson:"updatedAt"`
	ResolvedAt *time.Time    `bson:"resolvedAt,omitempty"`
}

// FeedbackDTO 是对外输出的反馈数据结构。
type FeedbackDTO struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	Contact    string     `json:"contact"`
	Project    string     `json:"project"`
	PageURL    string     `json:"pageUrl"`
	UserAgent  string     `json:"userAgent"`
	IP         string     `json:"ip"`
	Status     Status     `json:"status"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	ResolvedAt *time.Time `json:"resolvedAt,omitempty"`
}

// DTO 将文档转换为对外结构，避免直接暴露 bson.ObjectID。
func (f *Feedback) DTO() FeedbackDTO {
	return FeedbackDTO{
		ID:         f.ID.Hex(),
		Title:      f.Title,
		Content:    f.Content,
		Contact:    f.Contact,
		Project:    f.Project,
		PageURL:    f.PageURL,
		UserAgent:  f.UserAgent,
		IP:         f.IP,
		Status:     f.Status,
		CreatedAt:  f.CreatedAt,
		UpdatedAt:  f.UpdatedAt,
		ResolvedAt: f.ResolvedAt,
	}
}

// SubmitRequest 是公开提交接口的请求体，所有字段均可选。
type SubmitRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Contact string `json:"contact"`
	Project string `json:"project"`
	PageURL string `json:"pageUrl"`
}

// RequestMeta 是服务端自动采集的请求元数据。
type RequestMeta struct {
	IP        string
	UserAgent string
}

// Counts 是反馈总量的统计结果。
type Counts struct {
	Total    int64 `json:"total"`
	Open     int64 `json:"open"`
	Resolved int64 `json:"resolved"`
}

// SubmitResponse 是提交成功后的返回体。
type SubmitResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}
