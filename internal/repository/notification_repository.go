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

const notificationsCollection = "notifications"

// NotificationRepository handles notification data operations
type NotificationRepository struct {
	client *mongodb.MongoClient
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(client *mongodb.MongoClient) *NotificationRepository {
	return &NotificationRepository{client: client}
}

// EnsureIndexes creates necessary indexes for optimal query performance
func (r *NotificationRepository) EnsureIndexes(ctx context.Context) error {
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
				{Key: "tenantId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("tenant_type_created_idx"),
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("tenant_status_created_idx"),
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("tenant_type_status_created_idx"),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: 1},
			},
			Options: options.Index().SetName("status_created_idx"),
		},
		{
			Keys: bson.D{
				{Key: "idempotencyKey", Value: 1},
			},
			Options: options.Index().
				SetName("idempotency_key_idx").
				SetUnique(true).
				SetSparse(true), // Sparse index to allow null values
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "priority", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("tenant_priority_created_idx"),
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "category", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("tenant_category_created_idx"),
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "groupId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("tenant_group_created_idx"),
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "tags", Value: 1},
			},
			Options: options.Index().SetName("tenant_tags_idx"),
		},
		{
			Keys: bson.D{
				{Key: "scheduledFor", Value: 1},
			},
			Options: options.Index().
				SetName("scheduled_for_idx").
				SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "expiresAt", Value: 1},
			},
			Options: options.Index().
				SetName("expires_at_idx").
				SetSparse(true).
				SetExpireAfterSeconds(0), // TTL index
		},
	}

	return r.client.CreateIndexes(ctx, notificationsCollection, indexes)
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	notification.ID = primitive.NewObjectID()
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	_, err := r.client.Collection(notificationsCollection).InsertOne(ctx, notification)
	return err
}

// FindByID finds a notification by ID
func (r *NotificationRepository) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var notification domain.Notification
	err = r.client.Collection(notificationsCollection).FindOne(ctx, bson.M{"_id": objectID}).Decode(&notification)
	if err != nil {
		return nil, err
	}

	return &notification, nil
}

// Update updates a notification
func (r *NotificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	notification.UpdatedAt = time.Now()

	filter := bson.M{"_id": notification.ID}
	update := bson.M{"$set": notification}

	_, err := r.client.Collection(notificationsCollection).UpdateOne(ctx, filter, update)
	return err
}

// FindByTenantID finds notifications by tenant ID with pagination
// Uses aggregation pipeline for better performance with count
func (r *NotificationRepository) FindByTenantID(ctx context.Context, tenantID string, notificationType domain.NotificationType, status domain.NotificationStatus, page, pageSize int) ([]*domain.Notification, int64, error) {
	matchStage := bson.M{"tenantId": tenantID}

	if notificationType != "" {
		matchStage["type"] = notificationType
	}
	if status != "" {
		matchStage["status"] = status
	}

	// Calculate pagination
	skip := (page - 1) * pageSize

	// Use aggregation pipeline for efficient count + results in one query
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$facet", Value: bson.M{
			"metadata": bson.A{bson.M{"$count": "total"}},
			"data": bson.A{
				bson.M{"$sort": bson.M{"createdAt": -1}},
				bson.M{"$skip": skip},
				bson.M{"$limit": pageSize},
			},
		}}},
	}

	cursor, err := r.client.Collection(notificationsCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	type Result struct {
		Metadata []struct {
			Total int64 `bson:"total"`
		} `bson:"metadata"`
		Data []*domain.Notification `bson:"data"`
	}

	var results []Result
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	if len(results) == 0 || len(results[0].Data) == 0 {
		return []*domain.Notification{}, 0, nil
	}

	total := int64(0)
	if len(results[0].Metadata) > 0 {
		total = results[0].Metadata[0].Total
	}

	return results[0].Data, total, nil
}

// UpdateStatus updates the status of a notification
func (r *NotificationRepository) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus, errorMsg string, sentAt *time.Time) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	if errorMsg != "" {
		update["$set"].(bson.M)["error"] = errorMsg
	}

	if sentAt != nil {
		update["$set"].(bson.M)["sentAt"] = sentAt
	}

	filter := bson.M{"_id": objectID}
	_, err = r.client.Collection(notificationsCollection).UpdateOne(ctx, filter, update)
	return err
}

// IncrementRetryCount increments the retry count of a notification
func (r *NotificationRepository) IncrementRetryCount(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$inc": bson.M{"retryCount": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	_, err = r.client.Collection(notificationsCollection).UpdateOne(ctx, filter, update)
	return err
}

// CreateBatch creates multiple notifications in a single database operation
func (r *NotificationRepository) CreateBatch(ctx context.Context, notifications []*domain.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	now := time.Now()
	documents := make([]interface{}, len(notifications))
	for i, notification := range notifications {
		notification.ID = primitive.NewObjectID()
		notification.CreatedAt = now
		notification.UpdatedAt = now
		documents[i] = notification
	}

	_, err := r.client.Collection(notificationsCollection).InsertMany(ctx, documents)
	return err
}

// FindByIdempotencyKey finds a notification by idempotency key
func (r *NotificationRepository) FindByIdempotencyKey(ctx context.Context, idempotencyKey string) (*domain.Notification, error) {
	var notification domain.Notification
	filter := bson.M{"idempotencyKey": idempotencyKey}
	err := r.client.Collection(notificationsCollection).FindOne(ctx, filter).Decode(&notification)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// UpdateDeliveryStatus updates delivery status with timestamp
func (r *NotificationRepository) UpdateDeliveryStatus(ctx context.Context, id string, status domain.NotificationStatus, timestamp time.Time) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	// Set appropriate timestamp based on status
	switch status {
	case domain.NotificationStatusSent:
		update["$set"].(bson.M)["sentAt"] = timestamp
	case domain.NotificationStatusDelivered:
		update["$set"].(bson.M)["deliveredAt"] = timestamp
	case domain.NotificationStatusRead:
		update["$set"].(bson.M)["readAt"] = timestamp
	case domain.NotificationStatusClicked:
		update["$set"].(bson.M)["clickedAt"] = timestamp
	}

	filter := bson.M{"_id": objectID}
	_, err = r.client.Collection(notificationsCollection).UpdateOne(ctx, filter, update)
	return err
}

// FindByGroupID finds notifications by group ID
func (r *NotificationRepository) FindByGroupID(ctx context.Context, tenantID, groupID string, page, pageSize int) ([]*domain.Notification, int64, error) {
	filter := bson.M{
		"tenantId": tenantID,
		"groupId":  groupID,
	}

	skip := (page - 1) * pageSize

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

	cursor, err := r.client.Collection(notificationsCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	type Result struct {
		Metadata []struct {
			Total int64 `bson:"total"`
		} `bson:"metadata"`
		Data []*domain.Notification `bson:"data"`
	}

	var results []Result
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	if len(results) == 0 || len(results[0].Data) == 0 {
		return []*domain.Notification{}, 0, nil
	}

	total := int64(0)
	if len(results[0].Metadata) > 0 {
		total = results[0].Metadata[0].Total
	}

	return results[0].Data, total, nil
}

// FindByCategory finds notifications by category
func (r *NotificationRepository) FindByCategory(ctx context.Context, tenantID, category string, page, pageSize int) ([]*domain.Notification, int64, error) {
	filter := bson.M{
		"tenantId": tenantID,
		"category": category,
	}

	skip := (page - 1) * pageSize

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

	cursor, err := r.client.Collection(notificationsCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	type Result struct {
		Metadata []struct {
			Total int64 `bson:"total"`
		} `bson:"metadata"`
		Data []*domain.Notification `bson:"data"`
	}

	var results []Result
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	if len(results) == 0 || len(results[0].Data) == 0 {
		return []*domain.Notification{}, 0, nil
	}

	total := int64(0)
	if len(results[0].Metadata) > 0 {
		total = results[0].Metadata[0].Total
	}

	return results[0].Data, total, nil
}

// FindByTags finds notifications by tags
func (r *NotificationRepository) FindByTags(ctx context.Context, tenantID string, tags []string, page, pageSize int) ([]*domain.Notification, int64, error) {
	filter := bson.M{
		"tenantId": tenantID,
		"tags":     bson.M{"$in": tags},
	}

	skip := (page - 1) * pageSize

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

	cursor, err := r.client.Collection(notificationsCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	type Result struct {
		Metadata []struct {
			Total int64 `bson:"total"`
		} `bson:"metadata"`
		Data []*domain.Notification `bson:"data"`
	}

	var results []Result
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	if len(results) == 0 || len(results[0].Data) == 0 {
		return []*domain.Notification{}, 0, nil
	}

	total := int64(0)
	if len(results[0].Metadata) > 0 {
		total = results[0].Metadata[0].Total
	}

	return results[0].Data, total, nil
}
