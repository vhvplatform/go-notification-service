package repository

import (
	"context"
	"sync"
	"time"

	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/shared/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const templatesCollection = "email_templates"

// TemplateCache holds cached templates
type TemplateCache struct {
	templates map[string]*domain.EmailTemplate
	mu        sync.RWMutex
	ttl       time.Duration
	entries   map[string]time.Time
}

// NewTemplateCache creates a new template cache
func NewTemplateCache(ttl time.Duration) *TemplateCache {
	return &TemplateCache{
		templates: make(map[string]*domain.EmailTemplate),
		entries:   make(map[string]time.Time),
		ttl:       ttl,
	}
}

// Get retrieves a template from cache
func (c *TemplateCache) Get(key string) (*domain.EmailTemplate, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	template, exists := c.templates[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(c.entries[key]) > c.ttl {
		return nil, false
	}

	return template, true
}

// Set stores a template in cache
func (c *TemplateCache) Set(key string, template *domain.EmailTemplate) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.templates[key] = template
	c.entries[key] = time.Now()
}

// Invalidate removes a template from cache
func (c *TemplateCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.templates, key)
	delete(c.entries, key)
}

// TemplateRepository handles template data operations
type TemplateRepository struct {
	client *mongodb.MongoClient
	cache  *TemplateCache
}

// NewTemplateRepository creates a new template repository with caching
func NewTemplateRepository(client *mongodb.MongoClient) *TemplateRepository {
	return &TemplateRepository{
		client: client,
		cache:  NewTemplateCache(5 * time.Minute), // 5 minute cache TTL
	}
}

// EnsureIndexes creates necessary indexes for optimal query performance
func (r *TemplateRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "name", Value: 1},
			},
			Options: options.Index().SetName("tenant_name_idx").SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("tenant_created_idx"),
		},
	}

	return r.client.CreateIndexes(ctx, templatesCollection, indexes)
}

// Create creates a new template
func (r *TemplateRepository) Create(ctx context.Context, template *domain.EmailTemplate) error {
	template.ID = primitive.NewObjectID()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	_, err := r.client.Collection(templatesCollection).InsertOne(ctx, template)
	return err
}

// FindByID finds a template by ID with caching
func (r *TemplateRepository) FindByID(ctx context.Context, id string) (*domain.EmailTemplate, error) {
	// Check cache first
	cacheKey := "id:" + id
	if template, found := r.cache.Get(cacheKey); found {
		return template, nil
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var template domain.EmailTemplate
	err = r.client.Collection(templatesCollection).FindOne(ctx, bson.M{"_id": objectID}).Decode(&template)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache.Set(cacheKey, &template)

	return &template, nil
}

// FindByName finds a template by name and tenant ID with caching
func (r *TemplateRepository) FindByName(ctx context.Context, tenantID, name string) (*domain.EmailTemplate, error) {
	// Check cache first
	cacheKey := "tenant:" + tenantID + ":name:" + name
	if template, found := r.cache.Get(cacheKey); found {
		return template, nil
	}

	var template domain.EmailTemplate
	filter := bson.M{"tenant_id": tenantID, "name": name}
	err := r.client.Collection(templatesCollection).FindOne(ctx, filter).Decode(&template)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache.Set(cacheKey, &template)

	return &template, nil
}

// Update updates a template and invalidates cache
func (r *TemplateRepository) Update(ctx context.Context, template *domain.EmailTemplate) error {
	template.UpdatedAt = time.Now()

	filter := bson.M{"_id": template.ID}
	update := bson.M{"$set": template}

	_, err := r.client.Collection(templatesCollection).UpdateOne(ctx, filter, update)
	
	// Invalidate cache entries
	if err == nil {
		r.cache.Invalidate("id:" + template.ID.Hex())
		r.cache.Invalidate("tenant:" + template.TenantID + ":name:" + template.Name)
	}

	return err
}

// Delete deletes a template and invalidates cache
func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Get template first to invalidate cache properly
	var template domain.EmailTemplate
	if err := r.client.Collection(templatesCollection).FindOne(ctx, bson.M{"_id": objectID}).Decode(&template); err == nil {
		r.cache.Invalidate("id:" + id)
		r.cache.Invalidate("tenant:" + template.TenantID + ":name:" + template.Name)
	}

	_, err = r.client.Collection(templatesCollection).DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
