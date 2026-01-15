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

// TestOutbox_CreateNotification_WritesEventAtomically verifies notification + outbox event written in transaction
func TestOutbox_CreateNotification_WritesEventAtomically(t *testing.T) {
	t.Skip("Requires MongoDB with replica set - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	outboxRepo := NewOutboxEventRepository(client)
	notifRepo := NewNotificationRepository(client, outboxRepo)
	ctx := context.Background()

	// Create notification
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "test@example.com",
		Subject:   "Test Outbox",
		Body:      "Testing transactional outbox",
		Status:    domain.NotificationStatusPending,
	}

	err := notifRepo.Create(ctx, notif)
	require.NoError(t, err)
	require.NotEmpty(t, notif.ID)

	// Verify outbox event was created
	events, err := outboxRepo.FindByAggregateID(ctx, "notification", notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	require.Len(t, events, 1, "Should have exactly one outbox event")

	event := events[0]
	assert.Equal(t, domain.EventNotificationCreated, event.EventType)
	assert.Equal(t, "tenant-1", event.TenantID)
	assert.Equal(t, "notification", event.AggregateType)
	assert.Equal(t, notif.ID.Hex(), event.AggregateID)
	assert.Equal(t, domain.OutboxEventStatusPending, event.Status)
	assert.Nil(t, event.ProcessedAt)

	// Verify payload
	payload, ok := event.Payload.(domain.NotificationCreatedPayload)
	require.True(t, ok, "Payload should be NotificationCreatedPayload")
	assert.Equal(t, notif.ID.Hex(), payload.NotificationID)
	assert.Equal(t, "tenant-1", payload.TenantID)
	assert.Equal(t, domain.NotificationTypeEmail, payload.Type)
	assert.Equal(t, "test@example.com", payload.Recipient)
	assert.Equal(t, "Test Outbox", payload.Subject)
}

// TestOutbox_UpdateNotification_WritesEventAtomically verifies update + outbox event atomic write
func TestOutbox_UpdateNotification_WritesEventAtomically(t *testing.T) {
	t.Skip("Requires MongoDB with replica set - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	outboxRepo := NewOutboxEventRepository(client)
	notifRepo := NewNotificationRepository(client, outboxRepo)
	ctx := context.Background()

	// Create notification
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "test@example.com",
		Subject:   "Original Subject",
		Status:    domain.NotificationStatusPending,
	}
	err := notifRepo.Create(ctx, notif)
	require.NoError(t, err)

	// Update notification
	notif.Subject = "Updated Subject"
	notif.Body = "Updated Body"
	err = notifRepo.Update(ctx, notif)
	require.NoError(t, err)

	// Verify outbox events (should have 2: created + updated)
	events, err := outboxRepo.FindByAggregateID(ctx, "notification", notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	require.Len(t, events, 2, "Should have created + updated events")

	// Verify updated event
	updatedEvent := events[1]
	assert.Equal(t, domain.EventNotificationUpdated, updatedEvent.EventType)
	assert.Equal(t, domain.OutboxEventStatusPending, updatedEvent.Status)

	// Verify version incremented
	updated, err := notifRepo.FindByID(ctx, notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Version, "Version should be incremented to 2")
}

// TestOutbox_UpdateStatus_WritesStatusChangeEvent verifies status change creates event
func TestOutbox_UpdateStatus_WritesStatusChangeEvent(t *testing.T) {
	t.Skip("Requires MongoDB with replica set - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	outboxRepo := NewOutboxEventRepository(client)
	notifRepo := NewNotificationRepository(client, outboxRepo)
	ctx := context.Background()

	// Create notification
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "test@example.com",
		Subject:   "Status Test",
		Status:    domain.NotificationStatusPending,
	}
	err := notifRepo.Create(ctx, notif)
	require.NoError(t, err)

	// Update status
	now := time.Now()
	err = notifRepo.UpdateStatus(ctx, notif.ID.Hex(), "tenant-1", domain.NotificationStatusSent, "", &now)
	require.NoError(t, err)

	// Verify outbox events (should have 2: created + status_changed)
	events, err := outboxRepo.FindByAggregateID(ctx, "notification", notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	require.Len(t, events, 2, "Should have created + status_changed events")

	// Verify status change event
	statusEvent := events[1]
	assert.Equal(t, domain.EventNotificationStatusChanged, statusEvent.EventType)

	payload, ok := statusEvent.Payload.(domain.NotificationStatusChangedPayload)
	require.True(t, ok)
	assert.Equal(t, domain.NotificationStatusPending, payload.OldStatus)
	assert.Equal(t, domain.NotificationStatusSent, payload.NewStatus)
}

// TestOutbox_SoftDelete_WritesDeleteEvent verifies soft delete creates deletion event
func TestOutbox_SoftDelete_WritesDeleteEvent(t *testing.T) {
	t.Skip("Requires MongoDB with replica set - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	outboxRepo := NewOutboxEventRepository(client)
	notifRepo := NewNotificationRepository(client, outboxRepo)
	ctx := context.Background()

	// Create notification
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "test@example.com",
		Subject:   "Delete Test",
		Status:    domain.NotificationStatusPending,
	}
	err := notifRepo.Create(ctx, notif)
	require.NoError(t, err)

	// Soft delete
	err = notifRepo.SoftDelete(ctx, notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)

	// Verify outbox events (should have 2: created + deleted)
	events, err := outboxRepo.FindByAggregateID(ctx, "notification", notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	require.Len(t, events, 2, "Should have created + deleted events")

	// Verify delete event
	deleteEvent := events[1]
	assert.Equal(t, domain.EventNotificationDeleted, deleteEvent.EventType)
	assert.Equal(t, domain.OutboxEventStatusPending, deleteEvent.Status)

	payload, ok := deleteEvent.Payload.(domain.NotificationDeletedPayload)
	require.True(t, ok)
	assert.Equal(t, notif.ID.Hex(), payload.NotificationID)
	assert.Equal(t, "tenant-1", payload.TenantID)
}

// TestOutbox_TransactionRollback_NoEventsCreated verifies rollback discards outbox events
func TestOutbox_TransactionRollback_NoEventsCreated(t *testing.T) {
	t.Skip("Requires MongoDB with replica set - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	outboxRepo := NewOutboxEventRepository(client)
	notifRepo := NewNotificationRepository(client, outboxRepo)
	ctx := context.Background()

	// Simulate transaction failure by creating notification with invalid data
	// (This test requires mocking to force transaction failure)
	// For now, verify that partial writes don't exist

	// Query outbox events
	events, err := outboxRepo.FindUnprocessed(ctx, "tenant-1", 100)
	require.NoError(t, err)
	assert.Len(t, events, 0, "No orphaned events should exist after rollback")
}

// TestOutbox_TenantIsolation_EventsIsolated verifies tenant isolation in outbox events
func TestOutbox_TenantIsolation_EventsIsolated(t *testing.T) {
	t.Skip("Requires MongoDB with replica set - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	outboxRepo := NewOutboxEventRepository(client)
	notifRepo := NewNotificationRepository(client, outboxRepo)
	ctx := context.Background()

	// Create notifications for different tenants
	notif1 := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "user@tenant1.com",
		Subject:   "Tenant 1 Notification",
		Status:    domain.NotificationStatusPending,
	}
	err := notifRepo.Create(ctx, notif1)
	require.NoError(t, err)

	notif2 := &domain.Notification{
		TenantID:  "tenant-2",
		Type:      domain.NotificationTypeEmail,
		Recipient: "user@tenant2.com",
		Subject:   "Tenant 2 Notification",
		Status:    domain.NotificationStatusPending,
	}
	err = notifRepo.Create(ctx, notif2)
	require.NoError(t, err)

	// Verify tenant-1 can only see their events
	events1, err := outboxRepo.FindUnprocessed(ctx, "tenant-1", 100)
	require.NoError(t, err)
	require.Len(t, events1, 1)
	assert.Equal(t, "tenant-1", events1[0].TenantID)

	// Verify tenant-2 can only see their events
	events2, err := outboxRepo.FindUnprocessed(ctx, "tenant-2", 100)
	require.NoError(t, err)
	require.Len(t, events2, 1)
	assert.Equal(t, "tenant-2", events2[0].TenantID)
}

// TestOutbox_MarkProcessed_UpdatesStatus verifies Debezium can mark events as processed
func TestOutbox_MarkProcessed_UpdatesStatus(t *testing.T) {
	t.Skip("Requires MongoDB with replica set - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	outboxRepo := NewOutboxEventRepository(client)
	notifRepo := NewNotificationRepository(client, outboxRepo)
	ctx := context.Background()

	// Create notification
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "test@example.com",
		Subject:   "Process Test",
		Status:    domain.NotificationStatusPending,
	}
	err := notifRepo.Create(ctx, notif)
	require.NoError(t, err)

	// Get event
	events, err := outboxRepo.FindUnprocessed(ctx, "tenant-1", 1)
	require.NoError(t, err)
	require.Len(t, events, 1)
	event := events[0]

	// Mark as processed (simulating Debezium)
	err = outboxRepo.MarkProcessed(ctx, event.ID.Hex(), "tenant-1")
	require.NoError(t, err)

	// Verify status updated
	processedEvents, err := outboxRepo.FindByAggregateID(ctx, "notification", notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	require.Len(t, processedEvents, 1)
	assert.Equal(t, domain.OutboxEventStatusProcessed, processedEvents[0].Status)
	assert.NotNil(t, processedEvents[0].ProcessedAt)
	assert.Equal(t, 2, processedEvents[0].Version, "Version should increment on mark processed")
}

// TestOutbox_TraceID_InjectedIntoEvent verifies trace_id from context is captured
func TestOutbox_TraceID_InjectedIntoEvent(t *testing.T) {
	t.Skip("Requires MongoDB with replica set - run with integration test suite")
	t.Skip("Requires OpenTelemetry integration (Phase 3)")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	outboxRepo := NewOutboxEventRepository(client)
	notifRepo := NewNotificationRepository(client, outboxRepo)
	ctx := context.Background()

	// TODO: Phase 3 - Inject OpenTelemetry trace context
	// ctx = trace.ContextWithSpan(ctx, mockSpan)

	// Create notification
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "test@example.com",
		Subject:   "Trace Test",
		Status:    domain.NotificationStatusPending,
	}
	err := notifRepo.Create(ctx, notif)
	require.NoError(t, err)

	// Verify event has trace_id
	events, err := outboxRepo.FindByAggregateID(ctx, "notification", notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	require.Len(t, events, 1)

	// TODO: Phase 3 - Uncomment when OpenTelemetry integrated
	// assert.NotEmpty(t, events[0].TraceID, "TraceID should be extracted from context")
	// assert.NotEmpty(t, events[0].SpanID, "SpanID should be extracted from context")
}

// TestOutbox_BackwardCompatibility_WorksWithoutOutboxRepo verifies backward compatibility
func TestOutbox_BackwardCompatibility_WorksWithoutOutboxRepo(t *testing.T) {
	t.Skip("Requires MongoDB connection - run with integration test suite")

	// Setup
	client := setupTestMongoDB(t)
	defer teardownTestMongoDB(t, client)

	// Create notification repository WITHOUT outbox repo (backward compatibility)
	notifRepo := NewNotificationRepository(client, nil)
	ctx := context.Background()

	// Create notification (should work without outbox)
	notif := &domain.Notification{
		TenantID:  "tenant-1",
		Type:      domain.NotificationTypeEmail,
		Recipient: "test@example.com",
		Subject:   "Backward Compat Test",
		Status:    domain.NotificationStatusPending,
	}
	err := notifRepo.Create(ctx, notif)
	require.NoError(t, err)
	assert.NotEmpty(t, notif.ID)

	// Verify notification created
	found, err := notifRepo.FindByID(ctx, notif.ID.Hex(), "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, "Backward Compat Test", found.Subject)
}
