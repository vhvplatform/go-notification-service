# Debezium CDC Setup Guide - Outbox Pattern

**Service:** go-notification-service  
**Date:** January 12, 2026  
**Phase:** 2 - Transactional Outbox Pattern  
**Status:** Ready for Implementation

---

## 📋 Overview

This guide explains how to set up Debezium Change Data Capture (CDC) on the `outbox_events` collection to implement the Transactional Outbox pattern per Global Architecture Rules.

### Key Principles:
1. **Debezium ONLY CDC from `outbox_events`** - Never from business entities
2. **Atomic Writes** - Events written in same transaction as entity changes
3. **Distributed Tracing** - Events contain `trace_id` for OpenTelemetry correlation
4. **At-Least-Once Delivery** - Kafka consumers must be idempotent

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────┐
│  Notification Service                               │
│  ┌──────────────┐         ┌──────────────────┐    │
│  │ Create()     │ ──TXN──>│ notifications    │    │
│  │ Notification │         └──────────────────┘    │
│  └──────┬───────┘                                  │
│         │                  ┌──────────────────┐    │
│         └─────────TXN─────>│ outbox_events    │◄── Debezium CDC
│                             └──────────────────┘    │
└─────────────────────────────────────────────────────┘
                                      │
                                      │ CDC Stream
                                      ▼
                            ┌──────────────────┐
                            │  Kafka Topics    │
                            │  - notification.* │
                            │  - template.*     │
                            │  - scheduled.*    │
                            └──────────────────┘
                                      │
                                      ▼
                            ┌──────────────────┐
                            │  Consumers       │
                            │  - Analytics     │
                            │  - Notification  │
                            │  - Audit Log     │
                            └──────────────────┘
```

---

## 🚀 Step 1: Prerequisites

### Required Software:
- ✅ MongoDB 4.0+ (with Replica Set enabled)
- ✅ Kafka 2.8+
- ✅ Debezium MongoDB Connector 2.0+
- ✅ Kafka Connect (Distributed mode recommended)

### Verify MongoDB Replica Set:
```bash
# Connect to MongoDB
mongo

# Check replica set status
rs.status()

# Expected: replica set name and PRIMARY status
# If not configured, initialize replica set:
rs.initiate({
  _id: "rs0",
  members: [{ _id: 0, host: "localhost:27017" }]
})
```

### Verify Kafka:
```bash
# Check Kafka is running
kafka-topics.sh --bootstrap-server localhost:9092 --list

# Create outbox events topic (if not auto-created)
kafka-topics.sh --bootstrap-server localhost:9092 \
  --create \
  --topic notification.outbox.events \
  --partitions 3 \
  --replication-factor 1
```

---

## 🔧 Step 2: Install Debezium Connector

### Download Debezium MongoDB Connector:
```bash
# Download from Debezium website
curl -O https://repo1.maven.org/maven2/io/debezium/debezium-connector-mongodb/2.4.0.Final/debezium-connector-mongodb-2.4.0.Final-plugin.tar.gz

# Extract to Kafka Connect plugins directory
tar -xzf debezium-connector-mongodb-2.4.0.Final-plugin.tar.gz -C /path/to/kafka/plugins/
```

### Configure Kafka Connect:
Edit `connect-distributed.properties`:
```properties
# Kafka Connect settings
bootstrap.servers=localhost:9092
group.id=connect-cluster
key.converter=org.apache.kafka.connect.json.JsonConverter
value.converter=org.apache.kafka.connect.json.JsonConverter
key.converter.schemas.enable=false
value.converter.schemas.enable=false

# Plugin path (must include Debezium connector)
plugin.path=/path/to/kafka/plugins

# Offset storage
offset.storage.topic=connect-offsets
offset.storage.replication.factor=1

# Config storage
config.storage.topic=connect-configs
config.storage.replication.factor=1

# Status storage
status.storage.topic=connect-status
status.storage.replication.factor=1
```

### Start Kafka Connect:
```bash
connect-distributed.sh config/connect-distributed.properties
```

---

## 📝 Step 3: Configure Debezium Connector

### Create Connector Configuration:
Save as `debezium-notification-outbox.json`:

```json
{
  "name": "notification-outbox-connector",
  "config": {
    "connector.class": "io.debezium.connector.mongodb.MongoDbConnector",
    "mongodb.connection.string": "mongodb://localhost:27017/?replicaSet=rs0",
    "mongodb.user": "debezium_user",
    "mongodb.password": "${DEBEZIUM_PASSWORD}",
    
    "database.include.list": "notification_service",
    "collection.include.list": "notification_service.outbox_events",
    
    "topic.prefix": "notification",
    "topic.creation.default.replication.factor": 1,
    "topic.creation.default.partitions": 3,
    
    "capture.mode": "change_streams_update_full",
    
    "transforms": "outbox",
    "transforms.outbox.type": "io.debezium.transforms.outbox.EventRouter",
    "transforms.outbox.table.field.event.id": "_id",
    "transforms.outbox.table.field.event.key": "aggregateId",
    "transforms.outbox.table.field.event.type": "eventType",
    "transforms.outbox.table.field.event.payload": "payload",
    "transforms.outbox.table.field.event.timestamp": "createdAt",
    "transforms.outbox.route.topic.replacement": "${routedByValue}",
    "transforms.outbox.table.expand.json.payload": "true",
    
    "tombstones.on.delete": "false",
    
    "snapshot.mode": "initial",
    
    "mongodb.poll.interval.ms": 1000,
    
    "max.batch.size": 2048,
    "max.queue.size": 8192,
    
    "errors.tolerance": "all",
    "errors.log.enable": "true",
    "errors.log.include.messages": "true"
  }
}
```

### Deploy Connector:
```bash
# Register connector with Kafka Connect
curl -X POST http://localhost:8083/connectors \
  -H "Content-Type: application/json" \
  -d @debezium-notification-outbox.json

# Verify connector is running
curl http://localhost:8083/connectors/notification-outbox-connector/status
```

**Expected Output:**
```json
{
  "name": "notification-outbox-connector",
  "connector": {
    "state": "RUNNING",
    "worker_id": "connect-1:8083"
  },
  "tasks": [{
    "id": 0,
    "state": "RUNNING",
    "worker_id": "connect-1:8083"
  }]
}
```

---

## 🎯 Step 4: Kafka Topic Strategy

### Topic Naming Convention:
Events are routed to topics based on `eventType` field:

| Event Type                    | Kafka Topic                   | Consumers            |
| ----------------------------- | ----------------------------- | -------------------- |
| `notification.created`        | `notification.created`        | Analytics, Audit     |
| `notification.status_changed` | `notification.status_changed` | Monitoring, Webhooks |
| `notification.updated`        | `notification.updated`        | Audit                |
| `notification.deleted`        | `notification.deleted`        | Cleanup, Audit       |
| `template.created`            | `template.created`            | Cache Invalidation   |
| `template.updated`            | `template.updated`            | Cache Invalidation   |
| `template.deleted`            | `template.deleted`            | Cache Invalidation   |

### Pre-Create Topics (Recommended):
```bash
#!/bin/bash
EVENTS=(
  "notification.created"
  "notification.status_changed"
  "notification.updated"
  "notification.deleted"
  "template.created"
  "template.updated"
  "template.deleted"
  "scheduled_notification.created"
  "scheduled_notification.executed"
)

for event in "${EVENTS[@]}"; do
  kafka-topics.sh --bootstrap-server localhost:9092 \
    --create \
    --topic "$event" \
    --partitions 3 \
    --replication-factor 1 \
    --config retention.ms=604800000
done
```

---

## 🧪 Step 5: Test Event Flow

### 1. Create Test Notification:
```bash
curl -X POST http://localhost:8080/api/v1/notifications/email \
  -H "X-Tenant-ID: test-tenant" \
  -H "Content-Type: application/json" \
  -d '{
    "to": ["test@example.com"],
    "subject": "Test Outbox Event",
    "body": "This should create an outbox event"
  }'
```

### 2. Verify Outbox Event Created:
```bash
mongo notification_service --eval '
  db.outbox_events.find({
    eventType: "notification.created",
    status: "pending"
  }).pretty()
'
```

**Expected:**
```javascript
{
  "_id": ObjectId("..."),
  "tenantId": "test-tenant",
  "aggregateType": "notification",
  "aggregateId": "...",
  "eventType": "notification.created",
  "payload": {
    "notificationId": "...",
    "tenantId": "test-tenant",
    "type": "email",
    "recipient": "test@example.com",
    "subject": "Test Outbox Event",
    "status": "pending",
    "createdAt": ISODate("...")
  },
  "traceId": "",
  "spanId": "",
  "status": "pending",
  "version": 1,
  "createdAt": ISODate("...")
}
```

### 3. Verify Kafka Message:
```bash
kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic notification.created \
  --from-beginning \
  --max-messages 1
```

**Expected Output:**
```json
{
  "notificationId": "...",
  "tenantId": "test-tenant",
  "type": "email",
  "recipient": "test@example.com",
  "subject": "Test Outbox Event",
  "status": "pending",
  "createdAt": "2026-01-12T..."
}
```

### 4. Verify Event Marked as Processed:
```bash
# Check if Debezium marked event as processed (should be done by consumer)
mongo notification_service --eval '
  db.outbox_events.find({
    eventType: "notification.created",
    status: "processed"
  }).count()
'
```

---

## 🔍 Step 6: Monitoring & Observability

### Metrics to Monitor:

1. **Outbox Queue Size:**
   ```javascript
   db.outbox_events.count({ status: "pending" })
   ```
   - **Warning:** > 1000 pending events
   - **Critical:** > 10000 pending events

2. **Failed Events:**
   ```javascript
   db.outbox_events.count({ status: "failed" })
   ```

3. **Debezium Lag:**
   ```bash
   kafka-consumer-groups.sh \
     --bootstrap-server localhost:9092 \
     --group connect-notification-outbox-connector \
     --describe
   ```

4. **Event Processing Rate:**
   ```bash
   kafka-run-class.sh kafka.tools.GetOffsetShell \
     --broker-list localhost:9092 \
     --topic notification.created
   ```

### Prometheus Metrics (Optional):
```yaml
# Add to Prometheus config
- job_name: 'kafka-connect'
  static_configs:
    - targets: ['localhost:8083']
  metrics_path: '/metrics'
```

---

## 🧹 Step 7: Cleanup Job (CRITICAL)

**Problem:** Outbox events accumulate infinitely, causing storage issues.

**Solution:** Periodic cleanup of old processed events.

### Cleanup Script:
```javascript
// cleanup_outbox.js
// Run daily via cron: 0 2 * * * mongo notification_service < cleanup_outbox.js

var cutoffDate = new Date();
cutoffDate.setDate(cutoffDate.getDate() - 7); // Keep 7 days

var result = db.outbox_events.deleteMany({
  status: "processed",
  processedAt: { $lt: cutoffDate }
});

print("Deleted " + result.deletedCount + " old processed events");
```

### Cron Job:
```bash
# Edit crontab
crontab -e

# Add cleanup job (runs daily at 2 AM)
0 2 * * * mongo notification_service < /path/to/cleanup_outbox.js >> /var/log/outbox_cleanup.log 2>&1
```

### Alternative: TTL Index (MongoDB):
```javascript
// Create TTL index to auto-delete after 7 days
db.outbox_events.createIndex(
  { processedAt: 1 },
  { 
    expireAfterSeconds: 604800, // 7 days in seconds
    partialFilterExpression: { status: "processed" }
  }
);
```

---

## 🆘 Troubleshooting

### Issue 1: Connector Fails to Start

**Error:** `ReplicaSetNoPrimary` or `Connection refused`

**Solution:**
```bash
# 1. Verify MongoDB replica set
mongo --eval "rs.status()"

# 2. Check connectivity
telnet localhost 27017

# 3. Review connector logs
curl http://localhost:8083/connectors/notification-outbox-connector/status

# 4. Restart connector
curl -X POST http://localhost:8083/connectors/notification-outbox-connector/restart
```

### Issue 2: Events Not Appearing in Kafka

**Debug Steps:**
```bash
# 1. Check outbox events exist
mongo notification_service --eval 'db.outbox_events.count({ status: "pending" })'

# 2. Check connector status
curl http://localhost:8083/connectors/notification-outbox-connector/status

# 3. Check Kafka Connect logs
tail -f /var/log/kafka/connect.log | grep ERROR

# 4. Verify topic exists
kafka-topics.sh --bootstrap-server localhost:9092 --list | grep notification
```

### Issue 3: High Lag / Slow Processing

**Symptoms:** Thousands of pending events, consumers lagging

**Solutions:**
1. **Increase Connector Parallelism:**
   ```json
   {
     "tasks.max": 3,
     "max.batch.size": 4096
   }
   ```

2. **Add More Kafka Partitions:**
   ```bash
   kafka-topics.sh --bootstrap-server localhost:9092 \
     --alter \
     --topic notification.created \
     --partitions 6
   ```

3. **Scale Consumers:**
   - Deploy multiple consumer instances
   - Ensure consumer group has multiple members

### Issue 4: Duplicate Events

**Cause:** At-least-once delivery guarantee of Kafka

**Solution:** Make consumers idempotent:
```go
// Example idempotent consumer
func ProcessEvent(event *domain.OutboxEvent) error {
    // Check if already processed
    if r.isProcessed(event.ID) {
        log.Info("Event already processed, skipping", "eventId", event.ID)
        return nil
    }
    
    // Process event
    if err := r.handleEvent(event); err != nil {
        return err
    }
    
    // Mark as processed
    return r.markProcessed(event.ID)
}
```

---

## 📊 Performance Tuning

### Connector Settings:
```json
{
  "mongodb.poll.interval.ms": 500,
  "max.batch.size": 4096,
  "max.queue.size": 16384,
  "cursor.max.await.time.ms": 5000
}
```

### MongoDB Optimization:
```javascript
// Compound index for efficient polling
db.outbox_events.createIndex(
  { status: 1, createdAt: 1 },
  { name: "debezium_poll_idx" }
);

// Covered query for Debezium
db.outbox_events.createIndex(
  { 
    status: 1, 
    createdAt: 1, 
    eventType: 1, 
    aggregateId: 1, 
    payload: 1 
  }
);
```

---

## 🔒 Security Considerations

### 1. MongoDB Authentication:
```bash
# Create dedicated Debezium user
mongo admin --eval '
db.createUser({
  user: "debezium_user",
  pwd: "STRONG_PASSWORD",
  roles: [
    { role: "read", db: "notification_service" },
    { role: "read", db: "local" }
  ]
})
'
```

### 2. Encrypt Credentials:
```bash
# Use environment variables
export DEBEZIUM_PASSWORD="STRONG_PASSWORD"

# Or use Kafka Connect secrets
kafka-configs.sh --bootstrap-server localhost:9092 \
  --entity-type connector-configs \
  --entity-name notification-outbox-connector \
  --alter \
  --add-config 'mongodb.password=${file:/secrets/debezium-pass.txt}'
```

### 3. TLS/SSL:
```json
{
  "mongodb.ssl.enabled": "true",
  "mongodb.ssl.invalid.hostname.allowed": "false"
}
```

---

## 📚 References

- [Debezium MongoDB Connector Docs](https://debezium.io/documentation/reference/stable/connectors/mongodb.html)
- [Transactional Outbox Pattern](https://microservices.io/patterns/data/transactional-outbox.html)
- [Kafka Connect Configuration](https://kafka.apache.org/documentation/#connect)
- [Global Architecture Rules](../../docs/architecture/NEW_ARCHITECHTURE.md)

---

**Setup Guide Prepared By:** GitHub Copilot (Senior Technical Lead AI)  
**Review Status:** Ready for DevOps Implementation  
**Estimated Setup Time:** 2-3 hours
