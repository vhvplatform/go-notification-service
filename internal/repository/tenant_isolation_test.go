package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/shared/mongodb"
)

// TestTenantIsolation_Create verifies that notifications are created with correct tenant_id
func TestTenantIsolation_Create(t *testing.T) {
	t.Skip("Requires MongoDB connection - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	repo := NewNotificationRepository(client, nil) // No cache for testing
	ctx := context.Background()

	// Create notification for tenant-1
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "user@tenant1.com",
		Subject:   "Test Email",
		Body:      "Hello from tenant-1",
		Status:    domain.NotificationStatusPending,
	}

	err := repo.Create(ctx, notif)
	require.NoError(t, err)
	require.NotEmpty(t, notif.ID)
	assert.Equal(t, 1, notif.Version, "Version should be initialized to 1")
	assert.NotNil(t, notif.CreatedAt)
	assert.Nil(t, notif.DeletedAt)
}

// TestTenantIsolation_FindByID verifies cross-tenant access is prevented
func TestTenantIsolation_FindByID(t *testing.T) {
	t.Skip("Requires MongoDB connection - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	repo := NewNotificationRepository(client, nil)
	ctx := context.Background()

	// Create notification for tenant-1
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "user@tenant1.com",
		Subject:   "Secret Data",
		Body:      "Confidential information",
		Status:    domain.NotificationStatusPending,
	}
	err := repo.Create(ctx, notif)
	require.NoError(t, err)

	// Test 1: Same tenant can access
	found, err := repo.FindByID(ctx, notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, "Secret Data", found.Subject)

	// Test 2: Different tenant CANNOT access (CRITICAL SECURITY TEST)
	notFound, err := repo.FindByID(ctx, notif.ID.Hex(), "tenant-2")
	assert.Error(t, err, "Cross-tenant access should be denied")
	assert.Nil(t, notFound)
}

// TestTenantIsolation_FindByTenantID verifies listing returns only tenant's data
func TestTenantIsolation_FindByTenantID(t *testing.T) {
	t.Skip("Requires MongoDB connection - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	repo := NewNotificationRepository(client, nil)
	ctx := context.Background()

	// Create notifications for different tenants
	tenant1Notifs := []string{"notif-1-a", "notif-1-b", "notif-1-c"}
	for _, subject := range tenant1Notifs {
		err := repo.Create(ctx, &domain.Notification{
			TenantID:  "tenant-1",
			Type:      domain.NotificationTypeEmail,
			Recipient: "user@tenant1.com",
			Subject:   subject,
			Status:    domain.NotificationStatusPending,
		})
		require.NoError(t, err)
	}

	tenant2Notifs := []string{"notif-2-a", "notif-2-b"}
	for _, subject := range tenant2Notifs {
		err := repo.Create(ctx, &domain.Notification{
			TenantID:  "tenant-2",
			Type:      domain.NotificationTypeEmail,
			Recipient: "user@tenant2.com",
			Subject:   subject,
			Status:    domain.NotificationStatusPending,
		})
		require.NoError(t, err)
	}

	// Test: tenant-1 should only see 3 notifications
	results, total, err := repo.FindByTenantID(ctx, "tenant-1", 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, results, 3)
	for _, notif := range results {
		assert.Equal(t, "tenant-1", notif.TenantID, "Should only return tenant-1's data")
	}

	// Test: tenant-2 should only see 2 notifications
	results2, total2, err := repo.FindByTenantID(ctx, "tenant-2", 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total2)
	assert.Len(t, results2, 2)
}

// TestSoftDelete_NotReturnedInQueries verifies soft-deleted records are filtered
func TestSoftDelete_NotReturnedInQueries(t *testing.T) {
	t.Skip("Requires MongoDB connection - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	repo := NewNotificationRepository(client, nil)
	ctx := context.Background()

	// Create notification
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "user@example.com",
		Subject:   "Test Soft Delete",
		Status:    domain.NotificationStatusPending,
	}
	err := repo.Create(ctx, notif)
	require.NoError(t, err)

	// Verify it exists
	found, err := repo.FindByID(ctx, notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	assert.NotNil(t, found)

	// Soft delete
	err = repo.SoftDelete(ctx, notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)

	// Should NOT be returned by FindByID (deletedAt filter)
	notFound, err := repo.FindByID(ctx, notif.ID.Hex(), "tenant-1")
	assert.Error(t, err, "Soft-deleted record should not be returned")
	assert.Nil(t, notFound)

	// Should NOT appear in FindByTenantID listing
	results, total, err := repo.FindByTenantID(ctx, "tenant-1", 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total, "Soft-deleted records should be excluded from listing")
	assert.Len(t, results, 0)
}

// TestSoftDelete_Restore verifies soft-deleted records can be restored
func TestSoftDelete_Restore(t *testing.T) {
	t.Skip("Requires MongoDB connection - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	repo := NewNotificationRepository(client, nil)
	ctx := context.Background()

	// Create and soft delete notification
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "user@example.com",
		Subject:   "Test Restore",
		Status:    domain.NotificationStatusPending,
	}
	err := repo.Create(ctx, notif)
	require.NoError(t, err)

	err = repo.SoftDelete(ctx, notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)

	// Restore
	err = repo.Restore(ctx, notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)

	// Should now be accessible again
	found, err := repo.FindByID(ctx, notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Nil(t, found.DeletedAt, "DeletedAt should be cleared after restore")
}

// TestOptimisticLocking_ConcurrentUpdateConflict verifies version-based locking
func TestOptimisticLocking_ConcurrentUpdateConflict(t *testing.T) {
	t.Skip("Requires MongoDB connection - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	repo := NewNotificationRepository(client, nil)
	ctx := context.Background()

	// Create notification
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "user@example.com",
		Subject:   "Optimistic Lock Test",
		Status:    domain.NotificationStatusPending,
		Version:   1,
	}
	err := repo.Create(ctx, notif)
	require.NoError(t, err)

	// Simulate first update (succeeds)
	notif.Subject = "Updated by User A"
	err = repo.Update(ctx, notif)
	require.NoError(t, err)
	assert.Equal(t, 2, notif.Version, "Version should increment to 2")

	// Simulate concurrent update with OLD version (should fail)
	staleNotif := &domain.Notification{
		ID:        notif.ID,
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "user@example.com",
		Subject:   "Updated by User B (with stale version)",
		Status:    domain.NotificationStatusPending,
		Version:   1, // Stale version!
	}

	err = repo.Update(ctx, staleNotif)
	assert.Error(t, err, "Update with stale version should fail")
	assert.Contains(t, err.Error(), "concurrent modification", "Error should indicate optimistic lock conflict")
}

// TestUpdate_AutoIncrementVersion verifies version field is automatically incremented
func TestUpdate_AutoIncrementVersion(t *testing.T) {
	t.Skip("Requires MongoDB connection - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	repo := NewNotificationRepository(client, nil)
	ctx := context.Background()

	// Create notification
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "user@example.com",
		Subject:   "Version Test",
		Status:    domain.NotificationStatusPending,
	}
	err := repo.Create(ctx, notif)
	require.NoError(t, err)
	assert.Equal(t, 1, notif.Version)

	// Update 1
	notif.Subject = "Update 1"
	err = repo.Update(ctx, notif)
	require.NoError(t, err)
	assert.Equal(t, 2, notif.Version)

	// Update 2
	notif.Subject = "Update 2"
	err = repo.Update(ctx, notif)
	require.NoError(t, err)
	assert.Equal(t, 3, notif.Version)

	// Verify in DB
	found, err := repo.FindByID(ctx, notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, 3, found.Version)
}

// TestUpdate_AutoSetUpdatedAt verifies updatedAt is automatically set
func TestUpdate_AutoSetUpdatedAt(t *testing.T) {
	t.Skip("Requires MongoDB connection - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	repo := NewNotificationRepository(client, nil)
	ctx := context.Background()

	// Create notification
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "user@example.com",
		Subject:   "UpdatedAt Test",
		Status:    domain.NotificationStatusPending,
	}
	err := repo.Create(ctx, notif)
	require.NoError(t, err)
	originalUpdatedAt := notif.UpdatedAt

	// Wait to ensure time difference
	time.Sleep(100 * time.Millisecond)

	// Update
	notif.Subject = "Updated Subject"
	err = repo.Update(ctx, notif)
	require.NoError(t, err)

	// Verify updatedAt changed
	assert.True(t, notif.UpdatedAt.After(*originalUpdatedAt), "UpdatedAt should be refreshed on update")
}

// ============= Test Helpers =============

// setupTestMongoDB initializes a test MongoDB connection
func setupTestMongoDB(t *testing.T) *mongodb.MongoClient {
	// Use environment variable or default to local test instance
	// export MONGODB_TEST_URI="mongodb://localhost:27017/notification_service_test"
	uri := "mongodb://localhost:27017"
	database := "notification_service_test"

	config := &mongodb.Config{
		URI:            uri,
		Database:       database,
		ConnectTimeout: 10 * time.Second,
		MaxPoolSize:    10,
	}

	client, err := mongodb.NewMongoClient(config)
	require.NoError(t, err, "Failed to connect to test MongoDB")

	return client
}

// teardownTestMongoDB cleans up test database
func teardownTestMongoDB(t *testing.T, client *mongodb.MongoClient) {
	ctx := context.Background()

	// Drop test collections
	collections := []string{
		"notifications",
		"email_templates",
		"failed_notifications",
		"scheduled_notifications",
		"notification_preferences",
		"email_bounces",
	}

	for _, coll := range collections {
		err := client.Collection(coll).Drop(ctx)
		if err != nil {
			t.Logf("Warning: Failed to drop collection %s: %v", coll, err)
		}
	}

	client.Disconnect(ctx)
}
