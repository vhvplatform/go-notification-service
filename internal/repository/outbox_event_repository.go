package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/shared/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const outboxEventsCollection = "outbox_events"

// OutboxEventRepository handles outbox event data operations
// This repository is critical for Transactional Outbox Pattern (Phase 2)
type OutboxEventRepository struct {
	client *mongodb.MongoClient
}

// NewOutboxEventRepository creates a new outbox event repository
func NewOutboxEventRepository(client *mongodb.MongoClient) *OutboxEventRepository {
	return &OutboxEventRepository{client: client}
}

// EnsureIndexes creates necessary indexes for optimal query performance
func (r *OutboxEventRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: 1},
			},
			Options: options.Index().SetName("status_created_idx"),
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "aggregateType", Value: 1},
				{Key: "aggregateId", Value: 1},
			},
			Options: options.Index().SetName("tenant_aggregate_idx"),
		},
		{
			Keys: bson.D{
				{Key: "processedAt", Value: 1},
			},
			Options: options.Index().SetName("processed_at_idx").SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "traceId", Value: 1},
			},
			Options: options.Index().SetName("trace_id_idx").SetSparse(true),
		},
	}

	return r.client.CreateIndexes(ctx, outboxEventsCollection, indexes)
}

// Create creates a new outbox event
// This method should be called within the same transaction as the entity modification
func (r *OutboxEventRepository) Create(ctx context.Context, event *domain.OutboxEvent) error {
	event.ID = primitive.NewObjectID()
	event.Version = 1
	now := time.Now()
	event.CreatedAt = &now
	event.UpdatedAt = &now
	event.DeletedAt = nil

	// Default status
	if event.Status == "" {
		event.Status = domain.OutboxEventStatusPending
	}

	_, err := r.client.Collection(outboxEventsCollection).InsertOne(ctx, event)
	return err
}

// CreateWithSession creates a new outbox event within a MongoDB session (for transactions)
// CRITICAL: Use this method to ensure atomic writes with entity changes
func (r *OutboxEventRepository) CreateWithSession(ctx context.Context, session mongo.SessionContext, event *domain.OutboxEvent) error {
	event.ID = primitive.NewObjectID()
	event.Version = 1
	now := time.Now()
	event.CreatedAt = &now
	event.UpdatedAt = &now
	event.DeletedAt = nil

	// Default status
	if event.Status == "" {
		event.Status = domain.OutboxEventStatusPending
	}

	_, err := r.client.Collection(outboxEventsCollection).InsertOne(session, event)
	return err
}

// FindUnprocessed retrieves all pending events for processing by Debezium
func (r *OutboxEventRepository) FindUnprocessed(ctx context.Context, tenantID string, limit int) ([]*domain.OutboxEvent, error) {
	filter := bson.M{
		"tenantId":  tenantID,
		"status":    domain.OutboxEventStatusPending,
		"deletedAt": nil,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: 1}}).
		SetLimit(int64(limit))

	cursor, err := r.client.Collection(outboxEventsCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*domain.OutboxEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// MarkProcessed marks an outbox event as processed by Debezium
func (r *OutboxEventRepository) MarkProcessed(ctx context.Context, id string, tenantID string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":      domain.OutboxEventStatusProcessed,
			"processedAt": now,
			"updatedAt":   now,
		},
		"$inc": bson.M{"version": 1},
	}

	filter := bson.M{
		"_id":       objectID,
		"tenantId":  tenantID,
		"deletedAt": nil,
	}

	result, err := r.client.Collection(outboxEventsCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("outbox event not found or already deleted")
	}

	return nil
}

// MarkFailed marks an outbox event as failed with error details
func (r *OutboxEventRepository) MarkFailed(ctx context.Context, id string, tenantID string, errorMsg string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":    domain.OutboxEventStatusFailed,
			"lastError": errorMsg,
			"updatedAt": now,
		},
		"$inc": bson.M{
			"version":    1,
			"errorCount": 1,
		},
	}

	filter := bson.M{
		"_id":       objectID,
		"tenantId":  tenantID,
		"deletedAt": nil,
	}

	result, err := r.client.Collection(outboxEventsCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("outbox event not found or already deleted")
	}

	return nil
}

// FindByTraceID finds all events associated with a specific trace ID (for debugging)
func (r *OutboxEventRepository) FindByTraceID(ctx context.Context, traceID string, tenantID string) ([]*domain.OutboxEvent, error) {
	filter := bson.M{
		"traceId":   traceID,
		"tenantId":  tenantID,
		"deletedAt": nil,
	}

	cursor, err := r.client.Collection(outboxEventsCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*domain.OutboxEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// FindByAggregateID finds all events for a specific aggregate (e.g., all events for notification X)
func (r *OutboxEventRepository) FindByAggregateID(ctx context.Context, aggregateType string, aggregateID string, tenantID string) ([]*domain.OutboxEvent, error) {
	filter := bson.M{
		"aggregateType": aggregateType,
		"aggregateId":   aggregateID,
		"tenantId":      tenantID,
		"deletedAt":     nil,
	}

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}})

	cursor, err := r.client.Collection(outboxEventsCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*domain.OutboxEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// DeleteOldProcessedEvents cleans up old processed events (for maintenance)
// Should be run periodically to prevent outbox table bloat
func (r *OutboxEventRepository) DeleteOldProcessedEvents(ctx context.Context, olderThanDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	filter := bson.M{
		"status":      domain.OutboxEventStatusProcessed,
		"processedAt": bson.M{"$lt": cutoffDate},
	}

	result, err := r.client.Collection(outboxEventsCollection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}
