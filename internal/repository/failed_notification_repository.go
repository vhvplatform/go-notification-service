package repository

import (
	"context"
	"time"

	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/shared/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const failedNotificationsCollection = "failed_notifications"

// FailedNotificationRepository handles failed notification data operations
type FailedNotificationRepository struct {
	client *mongodb.MongoClient
}

// NewFailedNotificationRepository creates a new failed notification repository
func NewFailedNotificationRepository(client *mongodb.MongoClient) *FailedNotificationRepository {
	return &FailedNotificationRepository{client: client}
}

// EnsureIndexes creates necessary indexes for optimal query performance
func (r *FailedNotificationRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "failedAt", Value: -1},
			},
			Options: options.Index().SetName("tenant_failed_at_idx"),
		},
		{
			Keys: bson.D{
				{Key: "failedAt", Value: -1},
			},
			Options: options.Index().SetName("failed_at_idx"),
		},
	}

	return r.client.CreateIndexes(ctx, failedNotificationsCollection, indexes)
}

// Create creates a new failed notification record
func (r *FailedNotificationRepository) Create(ctx context.Context, failed *domain.FailedNotification) error {
	failed.ID = primitive.NewObjectID()
	failed.Version = 1
	failed.CreatedAt = time.Now()
	failed.UpdatedAt = time.Now()
	failed.DeletedAt = nil

	_, err := r.client.Collection(failedNotificationsCollection).InsertOne(ctx, failed)
	return err
}

// FindByID finds a failed notification by ID with tenant isolation
func (r *FailedNotificationRepository) FindByID(ctx context.Context, id string, tenantID string) (*domain.FailedNotification, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var failed domain.FailedNotification
	filter := bson.M{
		"_id":       objectID,
		"tenantId":  tenantID,
		"deletedAt": nil,
	}
	err = r.client.Collection(failedNotificationsCollection).FindOne(ctx, filter).Decode(&failed)
	if err != nil {
		return nil, err
	}

	return &failed, nil
}

// FindAll retrieves all failed notifications for a specific tenant with pagination
func (r *FailedNotificationRepository) FindAll(ctx context.Context, tenantID string, page, pageSize int) ([]*domain.FailedNotification, int64, error) {
	// Calculate pagination
	skip := (page - 1) * pageSize

	// Tenant isolation and soft delete filter
	matchStage := bson.M{
		"$match": bson.M{
			"tenantId":  tenantID,
			"deletedAt": nil,
		},
	}

	// Use aggregation pipeline for efficient count + results in one query
	pipeline := mongo.Pipeline{
		matchStage,
		{{Key: "$facet", Value: bson.M{
			"metadata": bson.A{bson.M{"$count": "total"}},
			"data": bson.A{
				bson.M{"$sort": bson.M{"failedAt": -1}},
				bson.M{"$skip": skip},
				bson.M{"$limit": pageSize},
			},
		}}},
	}

	cursor, err := r.client.Collection(failedNotificationsCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	type Result struct {
		Metadata []struct {
			Total int64 `bson:"total"`
		} `bson:"metadata"`
		Data []*domain.FailedNotification `bson:"data"`
	}

	var results []Result
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	if len(results) == 0 || len(results[0].Data) == 0 {
		return []*domain.FailedNotification{}, 0, nil
	}

	total := int64(0)
	if len(results[0].Metadata) > 0 {
		total = results[0].Metadata[0].Total
	}

	return results[0].Data, total, nil
}

// Delete deletes a failed notification by ID
func (r *FailedNotificationRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.client.Collection(failedNotificationsCollection).DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
