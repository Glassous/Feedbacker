// Package database 封装 MongoDB 连接与索引初始化。
package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"feedbacker/internal/config"
)

// Connect 建立 MongoDB 连接并验证可用性。
func Connect(cfg *config.Config, logger *slog.Logger) (*mongo.Client, *mongo.Database, error) {
	opts := options.Client().
		ApplyURI(cfg.MongoURI).
		SetAppName("feedbacker").
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second).
		SetMaxPoolSize(50).
		SetMinPoolSize(2)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, nil, fmt.Errorf("初始化 MongoDB 客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, fmt.Errorf("连接 MongoDB 失败: %w", err)
	}

	logger.Info("MongoDB 连接成功", "database", cfg.MongoDatabase)
	return client, client.Database(cfg.MongoDatabase), nil
}

// EnsureIndexes 创建查询所需的索引，可重复执行。
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	if _, err := db.Collection("feedback").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("idx_status_created_at"),
		},
		{
			Keys:    bson.D{{Key: "project", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("idx_project_created_at"),
		},
	}); err != nil {
		return fmt.Errorf("创建 feedback 索引失败: %w", err)
	}

	if _, err := db.Collection("admins").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetName("idx_admin_username").SetUnique(true),
	}); err != nil {
		return fmt.Errorf("创建 admins 索引失败: %w", err)
	}

	// 已处理反馈的过期清理等运营策略不在本期范围内，索引仅服务于查询。
	return nil
}
