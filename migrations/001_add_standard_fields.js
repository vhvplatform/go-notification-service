// MongoDB Migration Script - Add Standard Fields
// Date: 2026-01-12
// Purpose: Backfill version=1 and deletedAt=null for existing records
// Author: GitHub Copilot (Phase 1 Refactoring)

// Use notification service database
db = db.getSiblingDB('notification_service');

print('==============================================');
print('Phase 1 Migration: Adding Standard Fields');
print('Date: 2026-01-12');
print('==============================================\n');

// ========================================
// 1. Notifications Collection
// ========================================
print('Migrating notifications collection...');
var notificationsResult = db.notifications.updateMany(
    { version: { $exists: false } },
    {
        $set: {
            version: 1,
            deletedAt: null
        }
    }
);
print(`  - Updated ${notificationsResult.modifiedCount} documents`);
print(`  - Matched ${notificationsResult.matchedCount} documents\n`);

// Add indexes for deletedAt
print('Creating indexes for notifications...');
db.notifications.createIndex({ "tenantId": 1, "deletedAt": 1, "createdAt": -1 }, { name: "tenant_deleted_created_idx" });
db.notifications.createIndex({ "deletedAt": 1 }, { name: "deleted_at_idx", sparse: true });
print('  - Indexes created\n');

// ========================================
// 2. Email Templates Collection
// ========================================
print('Migrating email_templates collection...');
var templatesResult = db.email_templates.updateMany(
    { version: { $exists: false } },
    {
        $set: {
            version: 1,
            deletedAt: null
        }
    }
);
print(`  - Updated ${templatesResult.modifiedCount} documents`);
print(`  - Matched ${templatesResult.matchedCount} documents\n`);

// Add indexes
print('Creating indexes for email_templates...');
db.email_templates.createIndex({ "tenantId": 1, "deletedAt": 1, "name": 1 }, { name: "tenant_deleted_name_idx" });
db.email_templates.createIndex({ "deletedAt": 1 }, { name: "deleted_at_idx", sparse: true });
print('  - Indexes created\n');

// ========================================
// 3. Failed Notifications Collection
// ========================================
print('Migrating failed_notifications collection...');
var failedResult = db.failed_notifications.updateMany(
    { version: { $exists: false } },
    {
        $set: {
            version: 1,
            deletedAt: null,
            updatedAt: new Date() // Add missing updatedAt
        }
    }
);
print(`  - Updated ${failedResult.modifiedCount} documents`);
print(`  - Matched ${failedResult.matchedCount} documents\n`);

// Add indexes
print('Creating indexes for failed_notifications...');
db.failed_notifications.createIndex({ "tenantId": 1, "deletedAt": 1, "failedAt": -1 }, { name: "tenant_deleted_failed_idx" });
print('  - Indexes created\n');

// ========================================
// 4. Scheduled Notifications Collection
// ========================================
print('Migrating scheduled_notifications collection...');
var scheduledResult = db.scheduled_notifications.updateMany(
    { version: { $exists: false } },
    {
        $set: {
            version: 1,
            deletedAt: null
        }
    }
);
print(`  - Updated ${scheduledResult.modifiedCount} documents`);
print(`  - Matched ${scheduledResult.matchedCount} documents\n`);

// Add indexes
print('Creating indexes for scheduled_notifications...');
db.scheduled_notifications.createIndex({ "tenantId": 1, "deletedAt": 1, "isActive": 1 }, { name: "tenant_deleted_active_idx" });
db.scheduled_notifications.createIndex({ "deletedAt": 1 }, { name: "deleted_at_idx", sparse: true });
print('  - Indexes created\n');

// ========================================
// 5. Notification Preferences Collection
// ========================================
print('Migrating notification_preferences collection...');
var preferencesResult = db.notification_preferences.updateMany(
    { version: { $exists: false } },
    {
        $set: {
            version: 1,
            deletedAt: null
        }
    }
);
print(`  - Updated ${preferencesResult.modifiedCount} documents`);
print(`  - Matched ${preferencesResult.matchedCount} documents\n`);

// Add indexes
print('Creating indexes for notification_preferences...');
db.notification_preferences.createIndex({ "tenantId": 1, "deletedAt": 1, "userId": 1 }, { name: "tenant_deleted_user_idx" });
print('  - Indexes created\n');

// ========================================
// 6. Email Bounces Collection
// ========================================
print('Migrating email_bounces collection...');

// First, check if tenantId field exists
var sampleBounce = db.email_bounces.findOne();
if (sampleBounce && !sampleBounce.tenantId) {
    print('  ⚠️  WARNING: email_bounces collection is missing tenantId field!');
    print('  ⚠️  You need to manually assign tenant_id to existing bounces.');
    print('  ⚠️  Example: db.email_bounces.updateMany({tenantId: {$exists: false}}, {$set: {tenantId: "default-tenant"}});\n');
}

var bouncesResult = db.email_bounces.updateMany(
    { version: { $exists: false } },
    {
        $set: {
            version: 1,
            deletedAt: null,
            updatedAt: new Date() // Add missing updatedAt
        }
    }
);
print(`  - Updated ${bouncesResult.modifiedCount} documents`);
print(`  - Matched ${bouncesResult.matchedCount} documents\n`);

// Add indexes
print('Creating indexes for email_bounces...');
db.email_bounces.createIndex({ "tenantId": 1, "deletedAt": 1, "email": 1 }, { name: "tenant_deleted_email_idx" });
db.email_bounces.createIndex({ "deletedAt": 1 }, { name: "deleted_at_idx", sparse: true });
print('  - Indexes created\n');

// ========================================
// Summary
// ========================================
print('==============================================');
print('Migration Summary:');
print('==============================================');
print(`Notifications:             ${notificationsResult.modifiedCount} updated`);
print(`Email Templates:           ${templatesResult.modifiedCount} updated`);
print(`Failed Notifications:      ${failedResult.modifiedCount} updated`);
print(`Scheduled Notifications:   ${scheduledResult.modifiedCount} updated`);
print(`Notification Preferences:  ${preferencesResult.modifiedCount} updated`);
print(`Email Bounces:             ${bouncesResult.modifiedCount} updated`);
print('==============================================');
print('Migration completed successfully!');
print('==============================================\n');

// Verification queries
print('Verification Queries:');
print('---------------------------------------------');
print('1. Check notifications with new fields:');
print('   db.notifications.find({version: 1, deletedAt: null}).limit(1).pretty();\n');

print('2. Check soft-deleted records:');
print('   db.notifications.find({deletedAt: {$ne: null}}).count();\n');

print('3. Verify indexes:');
print('   db.notifications.getIndexes();\n');
