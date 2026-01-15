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
	client     *mongodb.MongoClient
	outboxRepo *OutboxEventRepository
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(client *mongodb.MongoClient, outboxRepo *OutboxEventRepository) *NotificationRepository {
	return &NotificationRepository{
		client:     client,
		outboxRepo: outboxRepo,
	}
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

// Create creates a new notification with transactional outbox event
// CRITICAL: Both notification and outbox event are written atomically
func (r *NotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	notification.ID = primitive.NewObjectID()
	notification.Version = 1
	now := time.Now()
	notification.CreatedAt = now
	notification.UpdatedAt = now
	notification.DeletedAt = nil

	// If outbox repository is not set, use simple insert (backward compatibility)
	if r.outboxRepo == nil {
		_, err := r.client.Collection(notificationsCollection).InsertOne(ctx, notification)
		return err
	}

	// Start MongoDB transaction for atomic write
	session, err := r.client.GetClient().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Execute transaction
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// 1. Insert notification
		_, err := r.client.Collection(notificationsCollection).InsertOne(sessCtx, notification)
		if err != nil {
			return nil, err
		}

		// 2. Create outbox event
		event := r.createNotificationCreatedEvent(ctx, notification)
		if err := r.outboxRepo.CreateWithSession(ctx, sessCtx, event); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

// FindByID finds a notification by ID with tenant isolation
func (r *NotificationRepository) FindByID(ctx context.Context, id string, tenantID string) (*domain.Notification, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var notification domain.Notification
	filter := bson.M{
		"_id":       objectID,
		"tenantId":  tenantID,
		"deletedAt": nil,
	}
	err = r.client.Collection(notificationsCollection).FindOne(ctx, filter).Decode(&notification)
	if err != nil {
		return nil, err
	}

	return &notification, nil
}

// Update updates a notification with tenant isolation and optimistic locking
func (r *NotificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	notification.UpdatedAt = time.Now()
	notification.Version++

	filter := bson.M{
		"_id":       notification.ID,
		"tenantId":  notification.TenantID,
		"deletedAt": nil,
		"version":   notification.Version - 1, // Optimistic locking
	}
	update := bson.M{"$set": notification}

	// If outbox repository is not set, use simple update (backward compatibility)
	if r.outboxRepo == nil {
		result, err := r.client.Collection(notificationsCollection).UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}
		if result.MatchedCount == 0 {
			return mongo.ErrNoDocuments
		}
		return nil
	}

	// Start MongoDB transaction for atomic write
	session, err := r.client.GetClient().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Execute transaction
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// 1. Update notification
		result, err := r.client.Collection(notificationsCollection).UpdateOne(sessCtx, filter, update)
		if err != nil {
			return nil, err
		}
		if result.MatchedCount == 0 {
			return nil, mongo.ErrNoDocuments
		}

		// 2. Create outbox event
		updatedFields := []string{"subject", "body", "metadata"} // TODO: Track actual changed fields
		event := r.createNotificationUpdatedEvent(ctx, notification, updatedFields)
		if err := r.outboxRepo.CreateWithSession(ctx, sessCtx, event); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

// FindByTenantID finds notifications by tenant ID with pagination
// Uses aggregation pipeline for better performance with count
func (r *NotificationRepository) FindByTenantID(ctx context.Context, tenantID string, notificationType domain.NotificationType, status domain.NotificationStatus, page, pageSize int) ([]*domain.Notification, int64, error) {
	matchStage := bson.M{
		"tenantId":  tenantID,
		"deletedAt": nil,
	}

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

// UpdateStatus updates the status of a notification with tenant isolation
func (r *NotificationRepository) UpdateStatus(ctx context.Context, id string, tenantID string, status domain.NotificationStatus, errorMsg string, sentAt *time.Time) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Fetch current notification to get old status
	var currentNotif domain.Notification
	if r.outboxRepo != nil {
		err := r.client.Collection(notificationsCollection).FindOne(ctx, bson.M{
			"_id":       objectID,
			"tenantId":  tenantID,
			"deletedAt": nil,
		}).Decode(&currentNotif)
		if err != nil {
			return err
		}
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
		"$inc": bson.M{"version": 1},
	}

	if errorMsg != "" {
		update["$set"].(bson.M)["error"] = errorMsg
	}

	if sentAt != nil {
		update["$set"].(bson.M)["sentAt"] = sentAt
	}

	filter := bson.M{
		"_id":       objectID,
		"tenantId":  tenantID,
		"deletedAt": nil,
	}

	// If outbox repository is not set, use simple update
	if r.outboxRepo == nil {
		_, err = r.client.Collection(notificationsCollection).UpdateOne(ctx, filter, update)
		return err
	}

	// Start MongoDB transaction
	session, err := r.client.GetClient().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Execute transaction
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// 1. Update status
		_, err := r.client.Collection(notificationsCollection).UpdateOne(sessCtx, filter, update)
		if err != nil {
			return nil, err
		}

		// 2. Create outbox event for status change
		currentNotif.Status = status
		currentNotif.UpdatedAt = time.Now()
		event := r.createNotificationStatusChangedEvent(ctx, &currentNotif, currentNotif.Status)
		if err := r.outboxRepo.CreateWithSession(ctx, sessCtx, event); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

// IncrementRetryCount increments the retry count of a notification with tenant isolation
func (r *NotificationRepository) IncrementRetryCount(ctx context.Context, id string, tenantID string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id":       objectID,
		"tenantId":  tenantID,
		"deletedAt": nil,
	}
	update := bson.M{
		"$inc": bson.M{
			"retryCount": 1,
			"version":    1,
		},
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
		notification.Version = 1
		notification.CreatedAt = now
		notification.UpdatedAt = now
		notification.DeletedAt = nil
		documents[i] = notification
	}

	_, err := r.client.Collection(notificationsCollection).InsertMany(ctx, documents)
	return err
}

// FindByIdempotencyKey finds a notification by idempotency key with tenant isolation
func (r *NotificationRepository) FindByIdempotencyKey(ctx context.Context, tenantID, idempotencyKey string) (*domain.Notification, error) {
	var notification domain.Notification
	filter := bson.M{
		"tenantId":       tenantID,
		"idempotencyKey": idempotencyKey,
		"deletedAt":      nil,
	}
	err := r.client.Collection(notificationsCollection).FindOne(ctx, filter).Decode(&notification)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// UpdateDeliveryStatus updates delivery status with timestamp and tenant isolation
func (r *NotificationRepository) UpdateDeliveryStatus(ctx context.Context, id string, tenantID string, status domain.NotificationStatus, timestamp time.Time) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
		"$inc": bson.M{"version": 1},
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

	filter := bson.M{
		"_id":       objectID,
		"tenantId":  tenantID,
		"deletedAt": nil,
	}
	_, err = r.client.Collection(notificationsCollection).UpdateOne(ctx, filter, update)
	return err
}

// FindByGroupID finds notifications by group ID with tenant isolation
func (r *NotificationRepository) FindByGroupID(ctx context.Context, tenantID, groupID string, page, pageSize int) ([]*domain.Notification, int64, error) {
	filter := bson.M{
		"tenantId":  tenantID,
		"groupId":   groupID,
		"deletedAt": nil,
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

// FindByCategory finds notifications by category with tenant isolation
func (r *NotificationRepository) FindByCategory(ctx context.Context, tenantID, category string, page, pageSize int) ([]*domain.Notification, int64, error) {
	filter := bson.M{
		"tenantId":  tenantID,
		"category":  category,
		"deletedAt": nil,
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

// FindByTags finds notifications by tags with tenant isolation
func (r *NotificationRepository) FindByTags(ctx context.Context, tenantID string, tags []string, page, pageSize int) ([]*domain.Notification, int64, error) {
	filter := bson.M{
		"tenantId":  tenantID,
		"tags":      bson.M{"$in": tags},
		"deletedAt": nil,
	}
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

// SoftDelete marks a notification as deleted (soft delete) with tenant isolation
func (r *NotificationRepository) SoftDelete(ctx context.Context, id string, tenantID string) error {
objectID, err := primitive.ObjectIDFromHex(id)
if err != nil {
return err
}

// Fetch notification before deletion to create event
var notification domain.Notification
if r.outboxRepo != nil {
	err := r.client.Collection(notificationsCollection).FindOne(ctx, bson.M{
		"_id":       objectID,
		"tenantId":  tenantID,
		"deletedAt": nil,
	}).Decode(&notification)
	if err != nil {
		return err
	}
}

now := time.Now()
filter := bson.M{
"_id":       objectID,
"tenantId":  tenantID,
"deletedAt": nil, // Only delete if not already deleted
}
update := bson.M{
"$set": bson.M{
"deletedAt": now,
"updatedAt": now,
},
"$inc": bson.M{"version": 1},
}

// If outbox repository is not set, use simple update
if r.outboxRepo == nil {
	result, err := r.client.Collection(notificationsCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// Start MongoDB transaction
session, err := r.client.GetClient().StartSession()
if err != nil {
	return err
}
defer session.EndSession(ctx)

// Execute transaction
_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
	// 1. Soft delete notification
	result, err := r.client.Collection(notificationsCollection).UpdateOne(sessCtx, filter, update)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, mongo.ErrNoDocuments
	}

	// 2. Create outbox event for deletion
	notification.DeletedAt = &now
	event := r.createNotificationDeletedEvent(ctx, &notification)
	if err := r.outboxRepo.CreateWithSession(ctx, sessCtx, event); err != nil {
		return nil, err
	}

	return nil, nil
})

return err
}

// Restore restores a soft-deleted notification with tenant isolation
func (r *NotificationRepository) Restore(ctx context.Context, id string, tenantID string) error {
objectID, err := primitive.ObjectIDFromHex(id)
if err != nil {
return err
}

filter := bson.M{
"_id":      objectID,
"tenantId": tenantID,
"deletedAt": bson.M{"$ne": nil}, // Only restore if deleted
}
update := bson.M{
"$set": bson.M{
"deletedAt": nil,
"updatedAt": time.Now(),
},
"$inc": bson.M{"version": 1},
}

result, err := r.client.Collection(notificationsCollection).UpdateOne(ctx, filter, update)
if err != nil {
return err
}
if result.MatchedCount == 0 {
return mongo.ErrNoDocuments
}
return nil
}

// ============= Outbox Event Helpers (Phase 2: Transactional Outbox) =============

// createNotificationCreatedEvent creates an outbox event for notification creation
func (r *NotificationRepository) createNotificationCreatedEvent(ctx context.Context, notification *domain.Notification) *domain.OutboxEvent {
	traceID, spanID := extractTraceContext(ctx)

	payload := domain.NotificationCreatedPayload{
		NotificationID: notification.ID.Hex(),
		TenantID:       notification.TenantID,
		Type:           notification.Type,
		Recipient:      notification.Recipient,
		Subject:        notification.Subject,
		Status:         notification.Status,
		CreatedAt:      notification.CreatedAt,
	}

	return &domain.OutboxEvent{
		TenantID:      notification.TenantID,
		AggregateType: "notification",
		AggregateID:   notification.ID.Hex(),
		EventType:     domain.EventNotificationCreated,
		Payload:       payload,
		TraceID:       traceID,
		SpanID:        spanID,
		Status:        domain.OutboxEventStatusPending,
	}
}

// createNotificationStatusChangedEvent creates an outbox event for status change
func (r *NotificationRepository) createNotificationStatusChangedEvent(ctx context.Context, notification *domain.Notification, oldStatus domain.NotificationStatus) *domain.OutboxEvent {
	traceID, spanID := extractTraceContext(ctx)

	payload := domain.NotificationStatusChangedPayload{
		NotificationID: notification.ID.Hex(),
		TenantID:       notification.TenantID,
		OldStatus:      oldStatus,
		NewStatus:      notification.Status,
		ChangedAt:      notification.UpdatedAt,
	}

	return &domain.OutboxEvent{
		TenantID:      notification.TenantID,
		AggregateType: "notification",
		AggregateID:   notification.ID.Hex(),
		EventType:     domain.EventNotificationStatusChanged,
		Payload:       payload,
		TraceID:       traceID,
		SpanID:        spanID,
		Status:        domain.OutboxEventStatusPending,
	}
}

// createNotificationUpdatedEvent creates an outbox event for notification update
func (r *NotificationRepository) createNotificationUpdatedEvent(ctx context.Context, notification *domain.Notification, updatedFields []string) *domain.OutboxEvent {
	traceID, spanID := extractTraceContext(ctx)

	payload := domain.NotificationUpdatedPayload{
		NotificationID: notification.ID.Hex(),
		TenantID:       notification.TenantID,
		Type:           notification.Type,
		UpdatedFields:  updatedFields,
		UpdatedAt:      notification.UpdatedAt,
	}

	return &domain.OutboxEvent{
		TenantID:      notification.TenantID,
		AggregateType: "notification",
		AggregateID:   notification.ID.Hex(),
		EventType:     domain.EventNotificationUpdated,
		Payload:       payload,
		TraceID:       traceID,
		SpanID:        spanID,
		Status:        domain.OutboxEventStatusPending,
	}
}

// createNotificationDeletedEvent creates an outbox event for notification deletion
func (r *NotificationRepository) createNotificationDeletedEvent(ctx context.Context, notification *domain.Notification) *domain.OutboxEvent {
	traceID, spanID := extractTraceContext(ctx)

	payload := domain.NotificationDeletedPayload{
		NotificationID: notification.ID.Hex(),
		TenantID:       notification.TenantID,
		DeletedAt:      *notification.DeletedAt,
	}

	return &domain.OutboxEvent{
		TenantID:      notification.TenantID,
		AggregateType: "notification",
		AggregateID:   notification.ID.Hex(),
		EventType:     domain.EventNotificationDeleted,
		Payload:       payload,
		TraceID:       traceID,
		SpanID:        spanID,
		Status:        domain.OutboxEventStatusPending,
	}
}

// extractTraceContext extracts OpenTelemetry trace ID and span ID from context
// CRITICAL: This enables distributed tracing across services via Kafka events
func extractTraceContext(ctx context.Context) (traceID string, spanID string) {
	// TODO: Implement OpenTelemetry trace extraction when OpenTelemetry is integrated (Phase 3)
	// For now, return empty strings (events will still be created, just without tracing)
	//
	// Example implementation (Phase 3):
	// span := trace.SpanFromContext(ctx)
	// if span.SpanContext().IsValid() {
	//     traceID = span.SpanContext().TraceID().String()
	//     spanID = span.SpanContext().SpanID().String()
	// }
	
	return "", ""
}
