package repository

import (
	"context"
	"time"

	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/shared/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const templatesCollection = "email_templates"

// TemplateRepository handles template data operations
type TemplateRepository struct {
	client *mongodb.MongoClient
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(client *mongodb.MongoClient) *TemplateRepository {
	return &TemplateRepository{client: client}
}

// Create creates a new template
func (r *TemplateRepository) Create(ctx context.Context, template *domain.EmailTemplate) error {
	template.ID = primitive.NewObjectID()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	_, err := r.client.Collection(templatesCollection).InsertOne(ctx, template)
	return err
}

// FindByID finds a template by ID
func (r *TemplateRepository) FindByID(ctx context.Context, id string) (*domain.EmailTemplate, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var template domain.EmailTemplate
	err = r.client.Collection(templatesCollection).FindOne(ctx, bson.M{"_id": objectID}).Decode(&template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

// FindByName finds a template by name and tenant ID
func (r *TemplateRepository) FindByName(ctx context.Context, tenantID, name string) (*domain.EmailTemplate, error) {
	var template domain.EmailTemplate
	filter := bson.M{"tenant_id": tenantID, "name": name}
	err := r.client.Collection(templatesCollection).FindOne(ctx, filter).Decode(&template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

// Update updates a template
func (r *TemplateRepository) Update(ctx context.Context, template *domain.EmailTemplate) error {
	template.UpdatedAt = time.Now()

	filter := bson.M{"_id": template.ID}
	update := bson.M{"$set": template}

	_, err := r.client.Collection(templatesCollection).UpdateOne(ctx, filter, update)
	return err
}

// Delete deletes a template
func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.client.Collection(templatesCollection).DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
