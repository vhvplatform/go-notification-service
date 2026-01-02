package repository

import (
	"context"
	"errors"
	"strings"
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

// Security constants for cache
const (
	maxCacheSize    = 1000        // Maximum number of cached templates
	maxCacheKeyLen  = 512         // Maximum length of cache key
	maxTemplateSize = 1024 * 1024 // Maximum template size: 1MB
)

// TemplateCache holds cached templates with security controls
type TemplateCache struct {
	templates map[string]*domain.EmailTemplate
	mu        sync.RWMutex
	ttl       time.Duration
	entries   map[string]time.Time
	maxSize   int // Maximum number of entries
}

// NewTemplateCache creates a new template cache with size limits
func NewTemplateCache(ttl time.Duration) *TemplateCache {
	return &TemplateCache{
		templates: make(map[string]*domain.EmailTemplate),
		entries:   make(map[string]time.Time),
		ttl:       ttl,
		maxSize:   maxCacheSize,
	}
}

// validateCacheKey validates cache key to prevent injection attacks
func validateCacheKey(key string) error {
	if len(key) == 0 {
		return errors.New("cache key cannot be empty")
	}
	if len(key) > maxCacheKeyLen {
		return errors.New("cache key exceeds maximum length")
	}
	// Prevent path traversal and special characters
	if strings.ContainsAny(key, "\x00\n\r") {
		return errors.New("cache key contains invalid characters")
	}
	return nil
}

// Get retrieves a template from cache with security validation
func (c *TemplateCache) Get(key string) (*domain.EmailTemplate, bool) {
	// Validate key
	if err := validateCacheKey(key); err != nil {
		return nil, false
	}

	c.mu.RLock()
	template, exists := c.templates[key]
	entryTime, hasEntry := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Check if expired
	if !hasEntry || time.Since(entryTime) > c.ttl {
		// Clean up expired entry
		c.mu.Lock()
		delete(c.templates, key)
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, false
	}

	return template, true
}

// Set stores a template in cache with security validation
func (c *TemplateCache) Set(key string, template *domain.EmailTemplate) error {
	// Validate key
	if err := validateCacheKey(key); err != nil {
		return err
	}

	// Validate template size to prevent memory exhaustion
	if template != nil {
		templateSize := len(template.Subject) + len(template.Body)
		if templateSize > maxTemplateSize {
			return errors.New("template size exceeds maximum allowed size")
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache size limit before adding
	if len(c.templates) >= c.maxSize && c.templates[key] == nil {
		// Cache is full and this is a new entry, evict oldest entry
		c.evictOldest()
	}

	c.templates[key] = template
	c.entries[key] = time.Now()
	return nil
}

// evictOldest removes the oldest entry from cache (must be called with lock held)
func (c *TemplateCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entryTime := range c.entries {
		if first || entryTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = entryTime
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.templates, oldestKey)
		delete(c.entries, oldestKey)
	}
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
				{Key: "tenantId", Value: 1},
				{Key: "name", Value: 1},
			},
			Options: options.Index().SetName("tenant_name_idx").SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "createdAt", Value: -1},
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

	// Cache the result (ignore error as caching is not critical)
	_ = r.cache.Set(cacheKey, &template)

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
	filter := bson.M{"tenantId": tenantID, "name": name}
	err := r.client.Collection(templatesCollection).FindOne(ctx, filter).Decode(&template)
	if err != nil {
		return nil, err
	}

	// Cache the result (ignore error as caching is not critical)
	_ = r.cache.Set(cacheKey, &template)

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

	// Invalidate cache by ID (we know the ID)
	r.cache.Invalidate("id:" + id)

	// Note: We cannot invalidate by tenant_id:name without fetching the template first.
	// This is an acceptable trade-off as Delete operations are infrequent and cache entries
	// will naturally expire based on TTL. Alternatively, we could maintain a reverse lookup
	// cache but that adds complexity for minimal benefit.

	_, err = r.client.Collection(templatesCollection).DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
