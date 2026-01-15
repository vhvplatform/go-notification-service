# Migration Guide - Phase 1: Standard Fields & Tenant Isolation

**Date:** January 12, 2026  
**Version:** 1.0  
**Status:** Ready for Production

---

## 📋 Overview

This migration adds critical standard fields to all collections and ensures tenant isolation compliance:
- `version` (int) - For optimistic locking
- `deletedAt` (timestamp) - For soft delete pattern
- `updatedAt` (timestamp) - Missing in some collections
- Indexes for performance optimization

---

## ⚠️ Pre-Migration Checklist

- [ ] **Backup Database**: Create full MongoDB backup
  ```bash
  mongodump --uri="mongodb://localhost:27017" --db=notification_service --out=/backup/$(date +%Y%m%d)
  ```

- [ ] **Stop Service**: Stop go-notification-service to prevent writes during migration
  ```bash
  docker-compose stop notification-service
  # OR
  systemctl stop go-notification-service
  ```

- [ ] **Verify MongoDB Connection**
  ```bash
  mongo --eval "db.adminCommand('ping')"
  ```

- [ ] **Check Disk Space**: Ensure sufficient space for index creation
  ```bash
  df -h /var/lib/mongodb
  ```

---

## 🚀 Migration Steps

### Step 1: Run Migration Script

```bash
cd /path/to/go-notification-service

# Connect to MongoDB and run migration
mongo notification_service < migrations/001_add_standard_fields.js

# Or using mongosh (MongoDB 5.0+)
mongosh notification_service < migrations/001_add_standard_fields.js
```

**Expected Output:**
```
==============================================
Phase 1 Migration: Adding Standard Fields
Date: 2026-01-12
==============================================

Migrating notifications collection...
  - Updated 1234 documents
  - Matched 1234 documents

Creating indexes for notifications...
  - Indexes created

... (similar for other collections)

==============================================
Migration completed successfully!
==============================================
```

### Step 2: Verify Migration

```bash
# Verify standard fields were added
mongo notification_service --eval '
db.notifications.findOne({}, {version: 1, deletedAt: 1, _id: 0})
'

# Expected: { "version": 1, "deletedAt": null }
```

### Step 3: Verify Indexes

```bash
# Check all indexes on notifications collection
mongo notification_service --eval 'db.notifications.getIndexes()'

# Expected: Should see tenant_deleted_created_idx, deleted_at_idx
```

---

## 🔧 Manual Steps Required

### ⚠️ CRITICAL: Assign Tenant IDs to Email Bounces

The `email_bounces` collection was missing `tenantId` field. You MUST manually assign tenant IDs:

```javascript
// Connect to MongoDB
use notification_service;

// Option 1: Set all bounces to default tenant
db.email_bounces.updateMany(
    { tenantId: { $exists: false } },
    { $set: { tenantId: "default-tenant-id" } }
);

// Option 2: Map bounces to tenants based on email domain
db.email_bounces.find({ tenantId: { $exists: false } }).forEach(function(bounce) {
    // Your logic to determine tenantId from email address
    var tenantId = determineTenantFromEmail(bounce.email);
    db.email_bounces.updateOne(
        { _id: bounce._id },
        { $set: { tenantId: tenantId } }
    );
});
```

---

## 📊 Performance Impact

### Index Creation Time (Estimated)

| Collection               | Documents | Index Time | Impact |
| ------------------------ | --------- | ---------- | ------ |
| notifications            | 100K      | ~30s       | Low    |
| email_templates          | 1K        | ~1s        | None   |
| failed_notifications     | 10K       | ~5s        | Low    |
| scheduled_notifications  | 500       | <1s        | None   |
| notification_preferences | 50K       | ~10s       | Low    |
| email_bounces            | 20K       | ~5s        | Low    |

**Total Downtime:** ~1-2 minutes for typical installation

---

## 🔄 Rollback Procedure

If migration fails or issues arise:

### Step 1: Restore from Backup

```bash
# Stop service first
docker-compose stop notification-service

# Restore database
mongorestore --uri="mongodb://localhost:27017" --db=notification_service /backup/20260112

# Verify restoration
mongo notification_service --eval 'db.notifications.count()'
```

### Step 2: Remove New Indexes (if needed)

```bash
mongo notification_service --eval '
db.notifications.dropIndex("tenant_deleted_created_idx");
db.notifications.dropIndex("deleted_at_idx");
// ... repeat for other collections
'
```

### Step 3: Revert Code

```bash
git revert <commit-hash>
docker-compose up -d --build notification-service
```

---

## ✅ Post-Migration Validation

### 1. Verify Application Startup

```bash
# Check logs for errors
docker-compose logs -f notification-service

# Should see: "Starting Notification Service..."
# Should NOT see: "Failed to connect to MongoDB"
```

### 2. Test Tenant Isolation

```bash
# Send test notification for tenant-1
curl -X POST http://localhost:8080/api/v1/notifications/email \
  -H "X-Tenant-ID: tenant-1" \
  -H "Content-Type: application/json" \
  -d '{
    "to": ["test@example.com"],
    "subject": "Test",
    "body": "Hello"
  }'

# Verify tenant-2 CANNOT see tenant-1's notification
curl http://localhost:8080/api/v1/notifications \
  -H "X-Tenant-ID: tenant-2"
  
# Expected: empty array []
```

### 3. Test Soft Delete

```bash
# Create notification
NOTIF_ID=$(curl -X POST ... | jq -r '.data.id')

# Soft delete it
curl -X DELETE http://localhost:8080/api/v1/notifications/$NOTIF_ID \
  -H "X-Tenant-ID: tenant-1"

# Verify it's not in listing (deletedAt IS NULL filter)
curl http://localhost:8080/api/v1/notifications \
  -H "X-Tenant-ID: tenant-1"

# Verify it's marked as deleted in DB
mongo notification_service --eval "
db.notifications.findOne(
  { _id: ObjectId('$NOTIF_ID') },
  { deletedAt: 1, _id: 0 }
)"

# Expected: { "deletedAt": ISODate("2026-01-12...") }
```

### 4. Test Optimistic Locking

```javascript
// In MongoDB shell
use notification_service;

// Get a notification
var notif = db.notifications.findOne();
print("Initial version:", notif.version);

// Update it
db.notifications.updateOne(
    { _id: notif._id, version: notif.version },
    { $set: { subject: "Updated" }, $inc: { version: 1 } }
);

// Try to update with old version (should fail)
var result = db.notifications.updateOne(
    { _id: notif._id, version: notif.version }, // Old version!
    { $set: { subject: "Updated Again" } }
);

print("Matched count:", result.matchedCount); // Expected: 0
```

---

## 📈 Monitoring

### Key Metrics to Watch

1. **Query Performance**
   ```bash
   # Check slow queries
   db.system.profile.find({ millis: { $gt: 100 } }).sort({ ts: -1 }).limit(10)
   ```

2. **Index Usage**
   ```bash
   db.notifications.aggregate([
     { $indexStats: {} }
   ])
   ```

3. **Collection Statistics**
   ```bash
   db.notifications.stats()
   ```

---

## 🆘 Troubleshooting

### Issue: Migration Script Fails

**Symptom:** Script exits with error
```
Error: E11000 duplicate key error collection
```

**Solution:**
```bash
# Check for duplicate indexes
db.notifications.getIndexes()

# Drop problematic index
db.notifications.dropIndex("index_name")

# Re-run migration
```

### Issue: Service Won't Start After Migration

**Symptom:** Logs show "version field not found"

**Solution:**
```bash
# Verify all documents have version field
db.notifications.count({ version: { $exists: false } })

# If count > 0, re-run migration
mongo notification_service < migrations/001_add_standard_fields.js
```

### Issue: Performance Degradation

**Symptom:** Queries slower after migration

**Solution:**
```bash
# Analyze query plans
db.notifications.find({ tenantId: "tenant-1", deletedAt: null }).explain("executionStats")

# Should show: "stage": "IXSCAN" (index scan, not COLLSCAN)
```

---

## 📝 Notes

- **Non-Blocking:** This migration uses `updateMany()` with small batches, safe for production
- **Idempotent:** Safe to run multiple times (checks `version: {$exists: false}`)
- **Backward Compatible:** Old code will ignore new fields initially

---

## 📞 Support

If you encounter issues:
1. Check logs: `docker-compose logs notification-service`
2. Verify database state: `mongo notification_service --eval 'db.notifications.findOne()'`
3. Contact DevOps team with error details

---

**Migration prepared by:** GitHub Copilot (Senior Technical Lead AI)  
**Review Status:** Ready for Production  
**Estimated Downtime:** 1-2 minutes
