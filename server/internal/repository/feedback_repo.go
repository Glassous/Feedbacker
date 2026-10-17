// Package repository 封装 MongoDB 数据访问。
package repository

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"feedbacker/internal/model"
)

// ErrNotFound 表示目标文档不存在。
var ErrNotFound = errors.New("记录不存在")

// FeedbackRepository 提供反馈文档的读写能力。
type FeedbackRepository struct {
	coll *mongo.Collection
}

// NewFeedbackRepository 创建反馈仓储。
func NewFeedbackRepository(db *mongo.Database) *FeedbackRepository {
	return &FeedbackRepository{coll: db.Collection("feedback")}
}

// ListParams 是列表查询条件。
type ListParams struct {
	Page     int64
	PageSize int64
	Query    string
	Status   string
	Project  string
}

// ListResult 是列表查询结果。
type ListResult struct {
	Items  []model.Feedback
	Total  int64
	Counts model.Counts
}

// Insert 写入一条反馈。
func (r *FeedbackRepository) Insert(ctx context.Context, f *model.Feedback) error {
	if f.ID == bson.NilObjectID {
		f.ID = bson.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, f)
	return err
}

// Get 按 ID 查询单条反馈。
func (r *FeedbackRepository) Get(ctx context.Context, id bson.ObjectID) (*model.Feedback, error) {
	var f model.Feedback
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&f)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// List 分页查询反馈，并返回全量状态统计。
func (r *FeedbackRepository) List(ctx context.Context, params ListParams) (*ListResult, error) {
	filter := buildFilter(params)

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip((params.Page - 1) * params.PageSize).
		SetLimit(params.PageSize)

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	items := make([]model.Feedback, 0, params.PageSize)
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	counts, err := r.Counts(ctx)
	if err != nil {
		return nil, err
	}

	return &ListResult{Items: items, Total: total, Counts: counts}, nil
}

// Counts 统计全部反馈的状态分布。
func (r *FeedbackRepository) Counts(ctx context.Context) (model.Counts, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$status"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return model.Counts{}, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		Status string `bson:"_id"`
		Count  int64  `bson:"count"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return model.Counts{}, err
	}

	var counts model.Counts
	for _, row := range rows {
		counts.Total += row.Count
		switch model.Status(row.Status) {
		case model.StatusOpen:
			counts.Open = row.Count
		case model.StatusResolved:
			counts.Resolved = row.Count
		}
	}
	return counts, nil
}

// UpdateStatus 更新反馈状态并返回更新后的文档。
func (r *FeedbackRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status model.Status) (*model.Feedback, error) {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": now,
		},
	}
	if status == model.StatusResolved {
		update["$set"].(bson.M)["resolvedAt"] = now
	} else {
		update["$unset"] = bson.M{"resolvedAt": ""}
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var f model.Feedback
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, opts).Decode(&f)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// Delete 删除一条反馈。
func (r *FeedbackRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// DistinctProjects 返回出现过的来源项目名称。
func (r *FeedbackRepository) DistinctProjects(ctx context.Context) ([]string, error) {
	var values []string
	if err := r.coll.Distinct(ctx, "project", bson.M{"project": bson.M{"$nin": bson.A{"", nil}}}).Decode(&values); err != nil {
		return nil, err
	}

	projects := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			projects = append(projects, value)
		}
	}
	sort.Strings(projects)
	return projects, nil
}

func buildFilter(params ListParams) bson.M {
	filter := bson.M{}

	if params.Status != "" {
		filter["status"] = params.Status
	}
	if params.Project != "" {
		filter["project"] = params.Project
	}
	if params.Query != "" {
		pattern := bson.M{"$regex": regexp.QuoteMeta(params.Query), "$options": "i"}
		filter["$or"] = bson.A{
			bson.M{"title": pattern},
			bson.M{"content": pattern},
			bson.M{"contact": pattern},
		}
	}

	return filter
}
