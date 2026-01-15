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

const scheduledNotificationsCollection = "scheduled_notifications"

// ScheduledNotificationRepository handles scheduled notification data operations
type ScheduledNotificationRepository struct {
	client *mongodb.MongoClient
}

// NewScheduledNotificationRepository creates a new repository
func NewScheduledNotificationRepository(client *mongodb.MongoClient) *ScheduledNotificationRepository {
	return &ScheduledNotificationRepository{client: client}
}

// EnsureIndexes creates necessary indexes for optimal query performance
func (r *ScheduledNotificationRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("tenant_created_idx"),
		},
		{
			Keys: bson.D{
				{Key: "isActive", Value: 1},
			},
			Options: options.Index().SetName("is_active_idx"),
		},
	}

	return r.client.CreateIndexes(ctx, scheduledNotificationsCollection, indexes)
}

// Create creates a new scheduled notification
func (r *ScheduledNotificationRepository) Create(ctx context.Context, scheduled *domain.ScheduledNotification) error {
	scheduled.ID = primitive.NewObjectID()
	scheduled.Version = 1
	scheduled.CreatedAt = time.Now()
	scheduled.UpdatedAt = time.Now()
	scheduled.DeletedAt = nil

	_, err := r.client.Collection(scheduledNotificationsCollection).InsertOne(ctx, scheduled)
	return err
}

// FindByID finds a scheduled notification by ID with tenant isolation
func (r *ScheduledNotificationRepository) FindByID(ctx context.Context, id string, tenantID string) (*domain.ScheduledNotification, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var scheduled domain.ScheduledNotification
	filter := bson.M{
		"_id":       objectID,
		"tenantId":  tenantID,
		"deletedAt": nil,
	}
	err = r.client.Collection(scheduledNotificationsCollection).FindOne(ctx, filter).Decode(&scheduled)
	if err != nil {
		return nil, err
	}

	return &scheduled, nil
}

// FindActive finds all active scheduled notifications (not soft deleted)
func (r *ScheduledNotificationRepository) FindActive(ctx context.Context) ([]*domain.ScheduledNotification, error) {
	filter := bson.M{
		"isActive":  true,
		"deletedAt": nil,
	}
	cursor, err := r.client.Collection(scheduledNotificationsCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var scheduled []*domain.ScheduledNotification
	if err = cursor.All(ctx, &scheduled); err != nil {
		return nil, err
	}

	return scheduled, nil
}

// FindByTenantID finds scheduled notifications by tenant ID with optimized pagination
func (r *ScheduledNotificationRepository) FindByTenantID(ctx context.Context, tenantID string, page, pageSize int) ([]*domain.ScheduledNotification, int64, error) {
	filter := bson.M{
		"tenantId":  tenantID,
		"deletedAt": nil,
	}

	// Calculate pagination
	skip := (page - 1) * pageSize

	// Use aggregation pipeline for efficient count + results in one query
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$facet", Value: bson.M{
			"metadata": bson.A{bson.M{"$count": "total"}},
			"data": bson.A{
				bson.M{"$sort": bson.M{"createdAt": -1}},
				bson.M{"$skip": skip},
				bson.M{"$limit": pageSize},
			},
		}}},
	}

	cursor, err := r.client.Collection(scheduledNotificationsCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	type Result struct {
		Metadata []struct {
			Total int64 `bson:"total"`
		} `bson:"metadata"`
		Data []*domain.ScheduledNotification `bson:"data"`
	}

	var results []Result
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	if len(results) == 0 || len(results[0].Data) == 0 {
		return []*domain.ScheduledNotification{}, 0, nil
	}

	total := int64(0)
	if len(results[0].Metadata) > 0 {
		total = results[0].Metadata[0].Total
	}

	return results[0].Data, total, nil
}

// Update updates a scheduled notification
func (r *ScheduledNotificationRepository) Update(ctx context.Context, scheduled *domain.ScheduledNotification) error {
	scheduled.UpdatedAt = time.Now()

	filter := bson.M{"_id": scheduled.ID}
	update := bson.M{"$set": scheduled}

	_, err := r.client.Collection(scheduledNotificationsCollection).UpdateOne(ctx, filter, update)
	return err
}

// Delete deletes a scheduled notification
func (r *ScheduledNotificationRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.client.Collection(scheduledNotificationsCollection).DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
