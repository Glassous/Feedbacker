package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Admin 是管理员账号文档，系统仅维护一个账号，不提供注册能力。
type Admin struct {
	ID                 bson.ObjectID `bson:"_id,omitempty"`
	Username           string        `bson:"username"`
	PasswordHash       string        `bson:"passwordHash"`
	MustChangePassword bool          `bson:"mustChangePassword"`
	TokenVersion       int           `bson:"tokenVersion"`
	CreatedAt          time.Time     `bson:"createdAt"`
	UpdatedAt          time.Time     `bson:"updatedAt"`
}

// AdminInfo 是可安全返回给前端的账号信息。
type AdminInfo struct {
	Username           string `json:"username"`
	MustChangePassword bool   `json:"mustChangePassword"`
}
