package repository

import (
	"context"
	"time"

	"github.com/longvhv/saas-framework-go/pkg/mongodb"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// Create creates a new failed notification record
func (r *FailedNotificationRepository) Create(ctx context.Context, failed *domain.FailedNotification) error {
	failed.ID = primitive.NewObjectID()
	failed.CreatedAt = time.Now()

	_, err := r.client.Collection(failedNotificationsCollection).InsertOne(ctx, failed)
	return err
}

// FindByID finds a failed notification by ID
func (r *FailedNotificationRepository) FindByID(ctx context.Context, id string) (*domain.FailedNotification, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var failed domain.FailedNotification
	err = r.client.Collection(failedNotificationsCollection).FindOne(ctx, bson.M{"_id": objectID}).Decode(&failed)
	if err != nil {
		return nil, err
	}

	return &failed, nil
}

// FindAll retrieves all failed notifications with pagination
func (r *FailedNotificationRepository) FindAll(ctx context.Context, page, pageSize int) ([]*domain.FailedNotification, int64, error) {
	// Get total count
	total, err := r.client.Collection(failedNotificationsCollection).CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	skip := (page - 1) * pageSize
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(bson.M{"failed_at": -1})

	cursor, err := r.client.Collection(failedNotificationsCollection).Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var failed []*domain.FailedNotification
	if err = cursor.All(ctx, &failed); err != nil {
		return nil, 0, err
	}

	return failed, total, nil
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
