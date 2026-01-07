# Server - Go Notification Service Backend

This directory contains the Golang backend microservice for the notification service.

## Structure

- `cmd/` - Main application entry point
- `internal/` - Internal packages
  - `consumer/` - Event consumers
  - `dlq/` - Dead Letter Queue handling
  - `domain/` - Domain models
  - `handler/` - HTTP handlers
  - `metrics/` - Metrics collection
  - `middleware/` - Middleware components
  - `queue/` - Queue management
  - `repository/` - Data access layer
  - `scheduler/` - Task scheduling
  - `service/` - Business logic services
  - `shared/` - Shared utilities
  - `smtp/` - SMTP handling
  - `webhook/` - Webhook management

## Building

```bash
make build
```

## Running

```bash
make run
```

## Testing

```bash
make test
```

For more details, see the Makefile in this directory.
