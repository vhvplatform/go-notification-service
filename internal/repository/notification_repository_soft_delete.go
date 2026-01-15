package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SoftDelete marks a notification as deleted (soft delete) with tenant isolation
func (r *NotificationRepository) SoftDelete(ctx context.Context, id string, tenantID string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
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

	result, err := r.client.Collection(notificationsCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// Restore restores a soft-deleted notification with tenant isolation
func (r *NotificationRepository) Restore(ctx context.Context, id string, tenantID string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id":       objectID,
		"tenantId":  tenantID,
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
