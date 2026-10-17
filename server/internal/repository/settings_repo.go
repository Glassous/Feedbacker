package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SettingsRepository 保存系统级配置（如 JWT 签名密钥）。
type SettingsRepository struct {
	coll *mongo.Collection
}

// NewSettingsRepository 创建设置仓储。
func NewSettingsRepository(db *mongo.Database) *SettingsRepository {
	return &SettingsRepository{coll: db.Collection("settings")}
}

// GetOrCreateJWTSecret 读取持久化的 JWT 密钥，不存在时生成并写入。
func (r *SettingsRepository) GetOrCreateJWTSecret(ctx context.Context) (string, error) {
	secret, err := r.findJWTSecret(ctx)
	if err == nil {
		return secret, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return "", err
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	generated := hex.EncodeToString(buf)

	_, err = r.coll.UpdateOne(ctx,
		bson.M{"_id": "jwt"},
		bson.M{"$setOnInsert": bson.M{"secret": generated, "createdAt": time.Now().UTC()}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return "", err
	}

	// 并发启动时可能由其他实例写入，重新读取以保持一致。
	return r.findJWTSecret(ctx)
}

func (r *SettingsRepository) findJWTSecret(ctx context.Context) (string, error) {
	var doc struct {
		Secret string `bson:"secret"`
	}
	err := r.coll.FindOne(ctx, bson.M{"_id": "jwt"}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if doc.Secret == "" {
		return "", ErrNotFound
	}
	return doc.Secret, nil
}
