// ==============================================
// Phase 2 Migration: Create Outbox Events Collection
// Date: 2026-01-12
// Purpose: Implement Transactional Outbox Pattern for Kafka event publishing
// ==============================================

print("==============================================");
print("Phase 2 Migration: Creating Outbox Events Collection");
print("Date: " + new Date().toISOString());
print("==============================================\n");

// Create outbox_events collection
print("Creating outbox_events collection...");
db.createCollection("outbox_events", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["tenantId", "aggregateType", "aggregateId", "eventType", "payload", "status", "version", "createdAt"],
            properties: {
                tenantId: {
                    bsonType: "string",
                    description: "Tenant ID for multi-tenancy isolation"
                },
                aggregateType: {
                    bsonType: "string",
                    enum: ["notification", "template", "scheduled_notification", "preference"],
                    description: "Type of aggregate (entity) that changed"
                },
                aggregateId: {
                    bsonType: "string",
                    description: "ID of the aggregate (entity) that changed"
                },
                eventType: {
                    bsonType: "string",
                    description: "Specific event type (e.g., notification.created)"
                },
                payload: {
                    bsonType: "object",
                    description: "Event payload with domain data"
                },
                traceId: {
                    bsonType: "string",
                    description: "OpenTelemetry trace ID for distributed tracing"
                },
                spanId: {
                    bsonType: "string",
                    description: "OpenTelemetry span ID for distributed tracing"
                },
                status: {
                    bsonType: "string",
                    enum: ["pending", "processed", "failed"],
                    description: "Processing status"
                },
                processedAt: {
                    bsonType: "date",
                    description: "Timestamp when Debezium processed this event"
                },
                errorCount: {
                    bsonType: "int",
                    description: "Number of failed processing attempts"
                },
                lastError: {
                    bsonType: "string",
                    description: "Last error message if processing failed"
                },
                version: {
                    bsonType: "int",
                    minimum: 1,
                    description: "Version for optimistic locking"
                },
                createdAt: {
                    bsonType: "date",
                    description: "Creation timestamp"
                },
                updatedAt: {
                    bsonType: "date",
                    description: "Last update timestamp"
                },
                deletedAt: {
                    bsonType: ["date", "null"],
                    description: "Soft delete timestamp"
                }
            }
        }
    }
});
print("✓ Collection created\n");

// Create indexes
print("Creating indexes for outbox_events...");

// 1. Status + CreatedAt index (for Debezium polling)
db.outbox_events.createIndex(
    { status: 1, createdAt: 1 },
    { name: "status_created_idx" }
);
print("  ✓ status_created_idx created");

// 2. Tenant + Aggregate index (for querying events by entity)
db.outbox_events.createIndex(
    { tenantId: 1, aggregateType: 1, aggregateId: 1 },
    { name: "tenant_aggregate_idx" }
);
print("  ✓ tenant_aggregate_idx created");

// 3. ProcessedAt index (sparse - for cleanup queries)
db.outbox_events.createIndex(
    { processedAt: 1 },
    { name: "processed_at_idx", sparse: true }
);
print("  ✓ processed_at_idx created");

// 4. TraceId index (sparse - for distributed tracing queries)
db.outbox_events.createIndex(
    { traceId: 1 },
    { name: "trace_id_idx", sparse: true }
);
print("  ✓ trace_id_idx created");

// 5. DeletedAt index (soft delete filtering)
db.outbox_events.createIndex(
    { deletedAt: 1 },
    { name: "deleted_at_idx", sparse: true }
);
print("  ✓ deleted_at_idx created");

// 6. Tenant + Status + CreatedAt (compound for tenant-specific polling)
db.outbox_events.createIndex(
    { tenantId: 1, status: 1, createdAt: 1 },
    { name: "tenant_status_created_idx" }
);
print("  ✓ tenant_status_created_idx created\n");

// Verify indexes
print("Verifying indexes...");
var indexes = db.outbox_events.getIndexes();
print("  Total indexes: " + indexes.length);
indexes.forEach(function (idx) {
    print("    - " + idx.name);
});

print("\n==============================================");
print("Phase 2 Migration completed successfully!");
print("==============================================\n");

print("⚠️  NEXT STEPS:");
print("1. Configure Debezium connector to CDC from 'outbox_events' collection ONLY");
print("2. Set up Kafka topics for event types (notification.*, template.*, etc.)");
print("3. Configure downstream consumers to process events");
print("4. Implement cleanup job for old processed events (see DEBEZIUM_SETUP.md)");
print("5. Enable OpenTelemetry trace injection (Phase 3)");
print("\nSee DEBEZIUM_SETUP.md for detailed configuration.\n");
