package repository

import (
	"context"
	"time"

	"github.com/vhvcorp/go-notification-service/internal/domain"
	"github.com/vhvcorp/go-notification-service/internal/shared/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// Create creates a new scheduled notification
func (r *ScheduledNotificationRepository) Create(ctx context.Context, scheduled *domain.ScheduledNotification) error {
	scheduled.ID = primitive.NewObjectID()
	scheduled.CreatedAt = time.Now()
	scheduled.UpdatedAt = time.Now()

	_, err := r.client.Collection(scheduledNotificationsCollection).InsertOne(ctx, scheduled)
	return err
}

// FindByID finds a scheduled notification by ID
func (r *ScheduledNotificationRepository) FindByID(ctx context.Context, id string) (*domain.ScheduledNotification, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var scheduled domain.ScheduledNotification
	err = r.client.Collection(scheduledNotificationsCollection).FindOne(ctx, bson.M{"_id": objectID}).Decode(&scheduled)
	if err != nil {
		return nil, err
	}

	return &scheduled, nil
}

// FindActive finds all active scheduled notifications
func (r *ScheduledNotificationRepository) FindActive(ctx context.Context) ([]*domain.ScheduledNotification, error) {
	filter := bson.M{"is_active": true}
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

// FindByTenantID finds scheduled notifications by tenant ID
func (r *ScheduledNotificationRepository) FindByTenantID(ctx context.Context, tenantID string, page, pageSize int) ([]*domain.ScheduledNotification, int64, error) {
	filter := bson.M{"tenant_id": tenantID}

	// Get total count
	total, err := r.client.Collection(scheduledNotificationsCollection).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	skip := (page - 1) * pageSize
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.client.Collection(scheduledNotificationsCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var scheduled []*domain.ScheduledNotification
	if err = cursor.All(ctx, &scheduled); err != nil {
		return nil, 0, err
	}

	return scheduled, total, nil
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
