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

const preferencesCollection = "notification_preferences"

// PreferencesRepository handles notification preferences data operations
type PreferencesRepository struct {
	client *mongodb.MongoClient
}

// NewPreferencesRepository creates a new preferences repository
func NewPreferencesRepository(client *mongodb.MongoClient) *PreferencesRepository {
	return &PreferencesRepository{client: client}
}

// EnsureIndexes creates necessary indexes for optimal query performance
func (r *PreferencesRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "userId", Value: 1},
			},
			Options: options.Index().SetName("tenant_user_idx").SetUnique(true),
		},
	}

	return r.client.CreateIndexes(ctx, preferencesCollection, indexes)
}

// GetByUserID retrieves preferences for a specific user with tenant isolation
func (r *PreferencesRepository) GetByUserID(ctx context.Context, tenantID, userID string) (*domain.NotificationPreferences, error) {
	var prefs domain.NotificationPreferences
	filter := bson.M{
		"tenantId":  tenantID,
		"userId":    userID,
		"deletedAt": nil,
	}
	err := r.client.Collection(preferencesCollection).FindOne(ctx, filter).Decode(&prefs)

	if err == mongo.ErrNoDocuments {
		// Return default preferences if not found
		return &domain.NotificationPreferences{
			TenantID:        tenantID,
			UserID:          userID,
			EmailEnabled:    true,
			SMSEnabled:      true,
			WebhookEnabled:  true,
			EmailCategories: make(map[string]bool),
			SMSCategories:   make(map[string]bool),
			Timezone:        "UTC",
		}, nil
	}

	return &prefs, err
}

// Create creates new preferences
func (r *PreferencesRepository) Create(ctx context.Context, prefs *domain.NotificationPreferences) error {
	prefs.ID = primitive.NewObjectID()
	prefs.Version = 1
	prefs.CreatedAt = time.Now()
	prefs.UpdatedAt = time.Now()
	prefs.DeletedAt = nil

	_, err := r.client.Collection(preferencesCollection).InsertOne(ctx, prefs)
	return err
}

// Update updates preferences with optimistic locking and tenant isolation
func (r *PreferencesRepository) Update(ctx context.Context, prefs *domain.NotificationPreferences) error {
	prefs.UpdatedAt = time.Now()
	prefs.Version++

	filter := bson.M{
		"tenantId":  prefs.TenantID,
		"userId":    prefs.UserID,
		"deletedAt": nil,
		"version":   prefs.Version - 1,
	}
	update := bson.M{"$set": prefs}
	opts := options.Update().SetUpsert(true)

	result, err := r.client.Collection(preferencesCollection).UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 && result.UpsertedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
