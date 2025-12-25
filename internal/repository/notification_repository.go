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

const notificationsCollection = "notifications"

// NotificationRepository handles notification data operations
type NotificationRepository struct {
	client *mongodb.MongoClient
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(client *mongodb.MongoClient) *NotificationRepository {
	return &NotificationRepository{client: client}
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
func (r *NotificationRepository) FindByTenantID(ctx context.Context, tenantID string, notificationType domain.NotificationType, status domain.NotificationStatus, page, pageSize int) ([]*domain.Notification, int64, error) {
	filter := bson.M{"tenant_id": tenantID}

	if notificationType != "" {
		filter["type"] = notificationType
	}
	if status != "" {
		filter["status"] = status
	}

	// Get total count
	total, err := r.client.Collection(notificationsCollection).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	skip := (page - 1) * pageSize
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.client.Collection(notificationsCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var notifications []*domain.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// UpdateStatus updates the status of a notification
func (r *NotificationRepository) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus, errorMsg string, sentAt *time.Time) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if errorMsg != "" {
		update["$set"].(bson.M)["error"] = errorMsg
	}

	if sentAt != nil {
		update["$set"].(bson.M)["sent_at"] = sentAt
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
		"$inc": bson.M{"retry_count": 1},
		"$set": bson.M{"updated_at": time.Now()},
	}

	_, err = r.client.Collection(notificationsCollection).UpdateOne(ctx, filter, update)
	return err
}
