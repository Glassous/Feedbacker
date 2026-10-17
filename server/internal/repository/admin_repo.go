package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"feedbacker/internal/model"
)

// AdminRepository 提供管理员账号的读写能力。
type AdminRepository struct {
	coll *mongo.Collection
}

// NewAdminRepository 创建管理员仓储。
func NewAdminRepository(db *mongo.Database) *AdminRepository {
	return &AdminRepository{coll: db.Collection("admins")}
}

// Count 返回管理员数量。
func (r *AdminRepository) Count(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{})
}

// FindByUsername 按用户名查询管理员。
func (r *AdminRepository) FindByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var admin model.Admin
	err := r.coll.FindOne(ctx, bson.M{"username": username}).Decode(&admin)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// FindByID 按 ID 查询管理员。
func (r *AdminRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.Admin, error) {
	var admin model.Admin
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&admin)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// Create 新建管理员账号。
func (r *AdminRepository) Create(ctx context.Context, admin *model.Admin) error {
	if admin.ID == bson.NilObjectID {
		admin.ID = bson.NewObjectID()
	}
	now := time.Now().UTC()
	admin.CreatedAt = now
	admin.UpdatedAt = now
	_, err := r.coll.InsertOne(ctx, admin)
	return err
}

// UpdatePassword 更新密码哈希，并使已签发的令牌失效。
func (r *AdminRepository) UpdatePassword(ctx context.Context, id bson.ObjectID, hash string, mustChange bool) error {
	update := bson.M{
		"$set": bson.M{
			"passwordHash":       hash,
			"mustChangePassword": mustChange,
			"updatedAt":          time.Now().UTC(),
		},
		"$inc": bson.M{"tokenVersion": 1},
	}
	result, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}
