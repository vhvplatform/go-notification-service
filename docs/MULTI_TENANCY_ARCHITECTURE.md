# Multi-Tenancy Architecture

## Overview

The `go-notification-service` has been aligned with the [go-infrastructure](https://github.com/vhvplatform/go-infrastructure) architectural standards to support a scalable multi-tenant SaaS platform.

## Architecture Alignment

### 1. Multi-Tenant Pattern

The service now implements tenant isolation using the **X-Tenant-ID header** pattern, consistent with the go-infrastructure standards:

- All `/api/v1/*` endpoints require the `X-Tenant-ID` header
- Tenant ID validation ensures proper format (3-128 characters)
- Tenant context is propagated through the entire request lifecycle
- Rate limiting is applied per-tenant for fair resource allocation

### 2. Tenancy Middleware

The service includes a tenancy middleware (`internal/middleware/tenancy.go`) that:

```go
// Extract and validate X-Tenant-ID header
tenantID := c.GetHeader(TenantIDHeader)

// Store in both Gin and standard contexts
c.Set(string(TenantIDKey), tenantID)
ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenantID)
```

**Key Features:**
- Header validation with error responses
- Context propagation for downstream services
- Audit logging support
- Helper functions for retrieving tenant ID

### 3. Infrastructure Integration

#### Namespace Strategy

Following go-infrastructure standards, the service deploys to the **`shared-services`** namespace:

```yaml
# Infrastructure namespaces:
# - core-system: Infrastructure components
# - shared-services: Platform services (Auth, Notifications, etc.)
# - tenant-workloads: Business microservices
# - 3rd-party-sandbox: Isolated 3rd party integrations
```

The notification service is classified as a shared platform service because it:
- Serves multiple tenants
- Provides common notification functionality
- Does not contain tenant-specific business logic

#### CI/CD Updates

All deployment workflows have been updated:
- Development: `shared-services` namespace
- Staging: `shared-services` namespace
- Production: `shared-services` namespace with blue-green deployment

### 4. Request Flow

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ X-Tenant-ID: tenant123
       ▼
┌─────────────┐
│   Ingress   │ (Pattern A: subfolder / Pattern B: custom domain)
└──────┬──────┘
       │ X-Tenant-ID: tenant123
       ▼
┌──────────────────────┐
│ Tenancy Middleware   │ Validates X-Tenant-ID
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│ Rate Limit Middleware│ Per-tenant rate limiting
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│   Handler Logic      │ Process notification
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│  Database/Queue      │ Tenant-isolated data
└──────────────────────┘
```

## API Changes

### Before
```bash
curl -X POST http://localhost:8084/api/v1/notifications/email \
  -H "Content-Type: application/json" \
  -d '{"tenant_id": "tenant123", ...}'
```

### After (Required)
```bash
curl -X POST http://localhost:8084/api/v1/notifications/email \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: tenant123" \
  -d '{"tenant_id": "tenant123", ...}'
```

**Note:** The X-Tenant-ID header is now **mandatory** for all `/api/v1/*` endpoints.

## Benefits

1. **Improved Security**: Explicit tenant identification prevents cross-tenant data access
2. **Better Observability**: Tenant context in logs and metrics
3. **Fair Resource Allocation**: Per-tenant rate limiting
4. **Architectural Consistency**: Aligned with platform-wide standards
5. **Scalability**: Supports hybrid multi-tenant patterns from go-infrastructure

## Migration Guide

### For API Consumers

1. **Add X-Tenant-ID Header**: All API requests must include the header
2. **Update Client Libraries**: Ensure SDKs/clients send the header
3. **Test Validation**: Verify error handling for missing/invalid tenant IDs

### For Infrastructure Teams

1. **Namespace Updates**: Deploy to `shared-services` namespace
2. **Network Policies**: Ensure network policies allow traffic to shared-services
3. **Monitoring**: Update dashboards to track per-tenant metrics

## Compatibility

### Breaking Changes
- All `/api/v1/*` endpoints now require X-Tenant-ID header
- Deployments moved to `shared-services` namespace

### Non-Breaking
- Health and metrics endpoints (`/health`, `/metrics`) remain accessible without headers
- Webhook endpoints for external providers don't require the header
- Existing functionality remains unchanged

## Related Documentation

- [go-infrastructure README](https://github.com/vhvplatform/go-infrastructure/blob/main/README.md)
- [Hybrid Multi-tenant Deployment Guide](https://github.com/vhvplatform/go-infrastructure/blob/main/docs/HYBRID_MULTITENANT_DEPLOYMENT.md)
- [Traffic Flow Architecture](https://github.com/vhvplatform/go-infrastructure/blob/main/docs/TRAFFIC_FLOW_ARCHITECTURE.md)

## Support

For questions about multi-tenancy:
- Review go-infrastructure documentation
- Check example implementations in services/middleware/
- Contact the platform team

## Future Enhancements

- [ ] Tenant-specific configuration overrides
- [ ] Advanced tenant routing patterns
- [ ] Tenant analytics and usage metrics
- [ ] Multi-region tenant data residency
