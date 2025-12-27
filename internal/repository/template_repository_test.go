package repository

import (
	"context"
	"testing"
	"time"

	"github.com/vhvplatform/go-notification-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestTemplateCache tests the template caching functionality
func TestTemplateCache(t *testing.T) {
	cache := NewTemplateCache(1 * time.Second)

	template := &domain.EmailTemplate{
		ID:       primitive.NewObjectID(),
		TenantID: "test-tenant",
		Name:     "test-template",
		Subject:  "Test Subject",
		Body:     "Test Body",
	}

	// Test Set and Get
	key := "test-key"
	cache.Set(key, template)

	retrieved, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached template")
	}
	if retrieved.Name != template.Name {
		t.Errorf("Expected template name %s, got %s", template.Name, retrieved.Name)
	}

	// Test cache expiration
	time.Sleep(1100 * time.Millisecond)
	_, found = cache.Get(key)
	if found {
		t.Error("Expected cache entry to be expired")
	}
}

// TestTemplateCacheInvalidate tests cache invalidation
func TestTemplateCacheInvalidate(t *testing.T) {
	cache := NewTemplateCache(5 * time.Minute)

	template := &domain.EmailTemplate{
		ID:       primitive.NewObjectID(),
		TenantID: "test-tenant",
		Name:     "test-template",
		Subject:  "Test Subject",
		Body:     "Test Body",
	}

	key := "test-key"
	cache.Set(key, template)

	// Verify it's cached
	_, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached template")
	}

	// Invalidate
	cache.Invalidate(key)

	// Verify it's removed
	_, found = cache.Get(key)
	if found {
		t.Error("Expected cache entry to be invalidated")
	}
}

// BenchmarkTemplateCacheGet benchmarks cache retrieval
func BenchmarkTemplateCacheGet(b *testing.B) {
	cache := NewTemplateCache(5 * time.Minute)

	template := &domain.EmailTemplate{
		ID:       primitive.NewObjectID(),
		TenantID: "test-tenant",
		Name:     "test-template",
		Subject:  "Test Subject",
		Body:     "Test Body",
	}

	cache.Set("test-key", template)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("test-key")
	}
}

// BenchmarkTemplateCacheSet benchmarks cache storage
func BenchmarkTemplateCacheSet(b *testing.B) {
	cache := NewTemplateCache(5 * time.Minute)

	template := &domain.EmailTemplate{
		ID:       primitive.NewObjectID(),
		TenantID: "test-tenant",
		Name:     "test-template",
		Subject:  "Test Subject",
		Body:     "Test Body",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("test-key", template)
	}
}

// TestFindByIDWithCache tests cached template retrieval
func TestFindByIDWithCache(t *testing.T) {
	t.Skip("Requires MongoDB connection - integration test")

	// This would test the caching layer in FindByID
	ctx := context.Background()
	_ = ctx
}

// TestFindByNameWithCache tests cached template retrieval by name
func TestFindByNameWithCache(t *testing.T) {
	t.Skip("Requires MongoDB connection - integration test")

	// This would test the caching layer in FindByName
	ctx := context.Background()
	_ = ctx
}
