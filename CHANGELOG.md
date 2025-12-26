# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Multi-tenant architecture support with X-Tenant-ID header validation
- Tenancy middleware aligned with go-infrastructure standards
- Integration with go-infrastructure architectural patterns
- Updated documentation with multi-tenant usage examples

### Changed
- Dockerfile updated to follow standard single-service structure
- CI/CD workflows now deploy to `shared-services` namespace per go-infrastructure standards
- All API endpoints now require X-Tenant-ID header for tenant isolation
- README updated with comprehensive multi-tenancy documentation

### Migration Notes
- All API requests to `/api/v1/*` endpoints now require the `X-Tenant-ID` header
- Deployments now use the `shared-services` Kubernetes namespace instead of environment-specific namespaces
- Service aligns with go-infrastructure multi-tenant architecture patterns

### Previous
- Initial extraction from monorepo
- CI/CD workflows
- Documentation

