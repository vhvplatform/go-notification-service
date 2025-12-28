# Performance Optimization Report

## Overview
This document describes the performance optimizations implemented in the Go Notification Service to improve database query efficiency, reduce latency, and enhance overall system throughput.

## Performance Bottlenecks Identified

### 1. Database Connection Pool
**Problem**: MongoDB client was using default connection pool settings, leading to connection exhaustion under high load.

**Solution**: Configured optimized connection pool settings:
- Max pool size: 100 connections
- Min pool size: 10 connections
- Max idle time: 30 seconds
- Connection timeout: 10 seconds
- Server selection timeout: 10 seconds

**Expected Impact**: 30-50% reduction in connection-related latency under concurrent load.

### 2. Missing Database Indexes
**Problem**: Queries by tenant_id, type, status, and created_at were performing full collection scans.

**Solution**: Added compound indexes for all repositories:
- `notifications`: 5 compound indexes covering common query patterns
- `email_templates`: 2 indexes including unique constraint on tenant_id + name
- `notification_preferences`: Unique index on tenant_id + user_id
- `scheduled_notifications`: 2 indexes for active status and tenant queries
- `email_bounces`: 2 compound indexes for email and type lookups
- `failed_notifications`: 2 indexes for failed_at queries

**Expected Impact**: 10-100x query performance improvement (depending on collection size).

### 3. Pagination Inefficiency (N+1 Query Problem)
**Problem**: Pagination required two separate database operations: `CountDocuments()` + `Find()`, leading to duplicate work.

**Solution**: Implemented single aggregation pipeline using `$facet` operator:
```javascript
{
  $facet: {
    metadata: [{ $count: "total" }],
    data: [
      { $sort: { created_at: -1 } },
      { $skip: skip },
      { $limit: pageSize }
    ]
  }
}
```

**Expected Impact**: 40-60% reduction in pagination query time.

### 4. Template Database Lookups
**Problem**: Every email sent required a database query to fetch templates, causing unnecessary database load.

**Solution**: Implemented in-memory LRU cache with:
- 5-minute TTL (configurable)
- Cache invalidation on template updates/deletes
- Thread-safe concurrent access with RWMutex

**Benchmark Results**:
- Cache Get: ~49 ns/op, 0 allocations
- Cache Set: ~93 ns/op, 0 allocations
- Database query: ~5-50 ms/op (100,000-1,000,000x slower)

**Expected Impact**: 95-99% reduction in template-related database queries.

### 5. Batch Insert Operations
**Problem**: Creating notifications one-by-one in loops caused N database round-trips.

**Solution**: Implemented batch insert using `InsertMany()`:
- `NotificationRepository.CreateBatch()`: Insert multiple notifications in one operation
- `EmailService.SendEmail()`: Now uses batch insert for multiple recipients

**Expected Impact**: 
- For 100 notifications: 99% reduction in database operations (100 → 1)
- For 1000 notifications: 99.9% reduction in database operations (1000 → 1)

### 6. Context Timeouts
**Problem**: Database operations could hang indefinitely, causing resource exhaustion.

**Solution**: Added 30-second context timeout to all worker operations in BulkEmailService.

**Expected Impact**: Improved system resilience and resource cleanup.

### 7. Template Variable Replacement
**Problem**: Multiple sequential `strings.ReplaceAll()` calls for each variable.

**Solution**: Optimized to use `strings.NewReplacer()` for batch replacement:
- Single pass through template string
- More efficient for templates with many variables

**Expected Impact**: 20-40% reduction in template rendering time for templates with 5+ variables.

## Performance Metrics

### Before Optimization (Estimated Baseline)
- Notification creation (100 emails): ~1000ms (10 ops each)
- Pagination query: ~100ms (2 operations)
- Template lookup: ~10ms per lookup
- Email sending throughput: ~500 emails/second

### After Optimization (Expected)
- Notification creation (100 emails): ~50ms (1 batch op)
- Pagination query: ~50ms (1 aggregation)
- Template lookup: ~0.00005ms (cached)
- Email sending throughput: ~1000-1500 emails/second

### Overall Expected Improvements
- Database operations: 60-95% reduction
- Query latency: 40-80% reduction (with indexes)
- Memory efficiency: Stable (cache with TTL and limits)
- System throughput: 2-3x increase
- CPU usage: 10-20% reduction (fewer DB round-trips)

## Testing and Validation

### Unit Tests Added
1. `TestTemplateCache` - Validates cache functionality and TTL
2. `TestTemplateCacheInvalidate` - Validates cache invalidation
3. `BenchmarkTemplateCacheGet` - Performance benchmark for cache reads
4. `BenchmarkTemplateCacheSet` - Performance benchmark for cache writes
5. `TestCreateBatch` - Validates batch insert (integration test)
6. `TestFindByTenantIDOptimized` - Validates aggregation pipeline (integration test)

### Benchmarks
```
BenchmarkTemplateCacheGet-4   	24245937	        49.30 ns/op	       0 B/op	       0 allocs/op
BenchmarkTemplateCacheSet-4   	12797361	        93.52 ns/op	       0 B/op	       0 allocs/op
```

### Integration Testing Required
To fully validate performance improvements in a real environment:
1. Load test with 1000+ concurrent requests
2. Measure database query times before/after indexes
3. Monitor cache hit rate (should be >90% for templates)
4. Profile memory usage under sustained load
5. Measure P95/P99 latency improvements

## Deployment Considerations

### Index Creation
Indexes should be created during deployment:
```go
// In main.go or initialization code
ctx := context.Background()

// Create indexes for all repositories
notifRepo.EnsureIndexes(ctx)
templateRepo.EnsureIndexes(ctx)
prefsRepo.EnsureIndexes(ctx)
scheduledRepo.EnsureIndexes(ctx)
bounceRepo.EnsureIndexes(ctx)
failedRepo.EnsureIndexes(ctx)
```

**Note**: Index creation on large existing collections may take time. Consider:
- Building indexes in background mode
- Scheduling during low-traffic periods
- Monitoring index build progress

### Cache Configuration
Template cache TTL can be adjusted based on requirements:
```go
// Current: 5 minutes
cache := NewTemplateCache(5 * time.Minute)

// For more frequently changing templates: 1 minute
cache := NewTemplateCache(1 * time.Minute)

// For rarely changing templates: 15 minutes
cache := NewTemplateCache(15 * time.Minute)
```

### Connection Pool Tuning
Connection pool sizes may need adjustment based on:
- Number of service instances
- Expected concurrent load
- MongoDB cluster capacity

Current settings (per instance):
- Max: 100 connections
- Min: 10 connections

For a cluster with 5 service instances:
- Total max connections: 500
- Ensure MongoDB can handle 500+ connections

## Monitoring and Metrics

### Key Metrics to Monitor
1. **Database Performance**
   - Query execution time (should decrease by 40-80%)
   - Connection pool utilization
   - Index usage statistics

2. **Cache Performance**
   - Template cache hit rate (target: >90%)
   - Cache size
   - Cache eviction rate

3. **Application Performance**
   - Email sending throughput (target: 2-3x increase)
   - P95/P99 latency (should decrease by 40-60%)
   - Failed notification rate (should remain stable)

4. **Resource Utilization**
   - CPU usage (should decrease by 10-20%)
   - Memory usage (should remain stable with cache)
   - Database connections (should not exhaust pool)

### Prometheus Metrics
Existing metrics to monitor:
```
notification_sent_total
notification_latency_seconds
smtp_pool_size
smtp_pool_available
queue_depth
```

Consider adding:
```
template_cache_hits_total
template_cache_misses_total
db_query_duration_seconds{operation="find|aggregate|insert"}
```

## Rollback Plan

If performance degradation is observed:

1. **Remove batch operations**: Revert to single inserts if batch operations cause issues
2. **Disable caching**: Set cache TTL to 0 or bypass cache layer
3. **Remove aggregation**: Revert to CountDocuments + Find pattern
4. **Drop indexes**: If indexes cause write performance issues (unlikely)
5. **Reduce pool size**: If connection exhaustion occurs at MongoDB level

## Recommendations for Further Optimization

### Short-term (Next Sprint)
1. Add cache metrics and monitoring
2. Implement cache warming on startup
3. Add database query timeout monitoring
4. Create load testing suite

### Medium-term (Next Quarter)
1. Implement Redis-based distributed cache (for multi-instance deployments)
2. Add query result caching for frequently accessed data
3. Implement read replicas for query distribution
4. Add database connection pooling per tenant

### Long-term (Next 6 Months)
1. Evaluate sharding strategy for large tenants
2. Implement event sourcing for audit trails
3. Consider CQRS pattern for read-heavy workloads
4. Evaluate materialized views for complex queries

## Conclusion

The implemented optimizations address the key performance bottlenecks in the notification service:
- Database connection pooling prevents resource exhaustion
- Strategic indexes dramatically improve query performance
- Aggregation pipelines reduce round-trips for pagination
- Template caching eliminates repeated database lookups
- Batch operations reduce database overhead

Expected overall impact: **2-3x throughput improvement** with **40-80% latency reduction** under typical production loads.

## References

- [MongoDB Performance Best Practices](https://www.mongodb.com/docs/manual/administration/analyzing-mongodb-performance/)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [MongoDB Aggregation Pipeline](https://www.mongodb.com/docs/manual/core/aggregation-pipeline/)
- [MongoDB Indexing Strategies](https://www.mongodb.com/docs/manual/applications/indexes/)
