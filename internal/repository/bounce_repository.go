package repository

import (
	"context"
	"time"

	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/shared/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const bouncesCollection = "email_bounces"

// BounceRepository handles email bounce data operations
type BounceRepository struct {
	client *mongodb.MongoClient
}

// NewBounceRepository creates a new bounce repository
func NewBounceRepository(client *mongodb.MongoClient) *BounceRepository {
	return &BounceRepository{client: client}
}

// Create creates a new bounce record
func (r *BounceRepository) Create(ctx context.Context, bounce *domain.EmailBounce) error {
	bounce.ID = primitive.NewObjectID()
	bounce.CreatedAt = time.Now()

	_, err := r.client.Collection(bouncesCollection).InsertOne(ctx, bounce)
	return err
}

// FindByEmail finds bounce records for an email address
func (r *BounceRepository) FindByEmail(ctx context.Context, email string) ([]*domain.EmailBounce, error) {
	filter := bson.M{"email": email}
	cursor, err := r.client.Collection(bouncesCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bounces []*domain.EmailBounce
	if err = cursor.All(ctx, &bounces); err != nil {
		return nil, err
	}

	return bounces, nil
}

// FindRecentHardBounces finds recent hard bounces for an email
func (r *BounceRepository) FindRecentHardBounces(ctx context.Context, email string, days int) ([]*domain.EmailBounce, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	filter := bson.M{
		"email":     email,
		"type":      "hard",
		"timestamp": bson.M{"$gte": cutoff},
	}

	cursor, err := r.client.Collection(bouncesCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bounces []*domain.EmailBounce
	if err = cursor.All(ctx, &bounces); err != nil {
		return nil, err
	}

	return bounces, nil
}
