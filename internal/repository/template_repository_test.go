package repository

import (
	"context"
	"fmt"
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
	_ = cache.Set(key, template)

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

// TestTemplateCacheSecurity tests security features of the cache
func TestTemplateCacheSecurity(t *testing.T) {
	cache := NewTemplateCache(5 * time.Minute)

	tests := []struct {
		name      string
		key       string
		template  *domain.EmailTemplate
		wantErr   bool
	}{
		{
			name: "valid key",
			key:  "valid-key",
			template: &domain.EmailTemplate{
				ID:      primitive.NewObjectID(),
				Subject: "Test",
				Body:    "Test",
			},
			wantErr: false,
		},
		{
			name:     "empty key",
			key:      "",
			template: &domain.EmailTemplate{},
			wantErr:  true,
		},
		{
			name:     "key too long",
			key:      string(make([]byte, 600)),
			template: &domain.EmailTemplate{},
			wantErr:  true,
		},
		{
			name:     "key with null byte",
			key:      "test\x00key",
			template: &domain.EmailTemplate{},
			wantErr:  true,
		},
		{
			name:     "key with newline",
			key:      "test\nkey",
			template: &domain.EmailTemplate{},
			wantErr:  true,
		},
		{
			name: "template too large",
			key:  "large-template",
			template: &domain.EmailTemplate{
				ID:      primitive.NewObjectID(),
				Subject: "Test",
				Body:    string(make([]byte, 2*1024*1024)), // 2MB
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set(tt.key, tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If set succeeded, verify Get also works
			if err == nil {
				_, found := cache.Get(tt.key)
				if !found {
					t.Error("Expected to find cached template after successful Set")
				}
			}
		})
	}
}

// TestTemplateCacheEviction tests that cache evicts oldest entries when full
func TestTemplateCacheEviction(t *testing.T) {
	cache := NewTemplateCache(1 * time.Minute)
	cache.maxSize = 5 // Set small size for testing

	// Fill cache to capacity
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key-%d", i)
		template := &domain.EmailTemplate{
			ID:      primitive.NewObjectID(),
			Subject: fmt.Sprintf("Subject %d", i),
			Body:    fmt.Sprintf("Body %d", i),
		}
		if err := cache.Set(key, template); err != nil {
			t.Fatalf("Failed to set key %s: %v", key, err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Verify cache is full
	if len(cache.templates) != 5 {
		t.Errorf("Expected cache size 5, got %d", len(cache.templates))
	}

	// Add one more - should evict oldest
	newKey := "key-new"
	template := &domain.EmailTemplate{
		ID:      primitive.NewObjectID(),
		Subject: "New Subject",
		Body:    "New Body",
	}
	if err := cache.Set(newKey, template); err != nil {
		t.Fatalf("Failed to set new key: %v", err)
	}

	// Verify cache size is still 5
	if len(cache.templates) != 5 {
		t.Errorf("Expected cache size 5 after eviction, got %d", len(cache.templates))
	}

	// Verify new key is present
	if _, found := cache.Get(newKey); !found {
		t.Error("Expected to find new key after adding to full cache")
	}

	// Verify oldest key (key-0) was evicted
	if _, found := cache.Get("key-0"); found {
		t.Error("Expected oldest key to be evicted")
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
