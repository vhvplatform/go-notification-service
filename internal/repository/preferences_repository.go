package repository

import (
	"context"
	"time"

	"github.com/vhvcorp/go-notification-service/internal/domain"
	"github.com/vhvcorp/go-shared/mongodb"
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

// GetByUserID retrieves preferences for a specific user
func (r *PreferencesRepository) GetByUserID(ctx context.Context, tenantID, userID string) (*domain.NotificationPreferences, error) {
	var prefs domain.NotificationPreferences
	filter := bson.M{"tenant_id": tenantID, "user_id": userID}
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
	prefs.CreatedAt = time.Now()
	prefs.UpdatedAt = time.Now()

	_, err := r.client.Collection(preferencesCollection).InsertOne(ctx, prefs)
	return err
}

// Update updates preferences
func (r *PreferencesRepository) Update(ctx context.Context, prefs *domain.NotificationPreferences) error {
	prefs.UpdatedAt = time.Now()
	filter := bson.M{"tenant_id": prefs.TenantID, "user_id": prefs.UserID}
	update := bson.M{"$set": prefs}
	opts := options.Update().SetUpsert(true)

	_, err := r.client.Collection(preferencesCollection).UpdateOne(ctx, filter, update, opts)
	return err
}
