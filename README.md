# Go Notification Service

> Multi-channel notification service supporting Email, SMS, Push Notifications, and Webhooks

## Description

A robust, scalable notification service built in Go that provides unified API for sending notifications across multiple channels. Features include template management, delivery tracking, retry mechanisms, rate limiting, and comprehensive monitoring.

## Features

### Multi-Channel Support
- **Email**: SMTP, SendGrid, AWS SES
- **SMS**: Twilio, AWS SNS
- **Push Notifications**: FCM (Android), APNs (iOS)
- **Webhooks**: Custom HTTP endpoints

### Core Capabilities
- **Template Management**: Reusable templates with variable substitution
- **Delivery Tracking**: Real-time status tracking and history
- **Retry Mechanism**: Automatic retry with exponential backoff
- **Dead Letter Queue**: Failed notification handling
- **Batch Processing**: Bulk notification sending
- **Scheduled Notifications**: Cron-based scheduling
- **User Preferences**: Per-user notification preferences
- **Rate Limiting**: Per-tenant rate limiting
- **Bounce Handling**: Email bounce tracking and management

### Reliability Features
- **Circuit Breaker**: Prevent cascading failures
- **Connection Pooling**: SMTP connection pool management
- **Idempotency**: Duplicate notification prevention
- **Graceful Degradation**: Fallback providers
- **Health Checks**: Service health monitoring

### Observability
- **Prometheus Metrics**: Performance and delivery metrics
- **Structured Logging**: JSON logging for easy parsing
- **Distributed Tracing**: Request tracing support
- **Error Tracking**: Detailed error reporting

## Prerequisites

- Go 1.25+
- MongoDB 4.4+
- RabbitMQ 3.9+
- Docker & Docker Compose (for containerized deployment)

## Quick Start

```bash
# Clone the repository
git clone https://github.com/vhvcorp/go-notification-service.git
cd go-notification-service

# Install dependencies
go mod download

# Set up environment variables
cp .env.example .env
# Edit .env with your configuration

# Run with Docker Compose (recommended for development)
docker-compose up -d

# Or run locally
make run
```

## Installation

```bash
# Clone the repository
git clone https://github.com/vhvcorp/go-notification-service.git
cd go-notification-service

# Install dependencies
go mod download
```

## Configuration

### Environment Variables

```bash
# Server Configuration
NOTIFICATION_SERVICE_PORT=8084

# MongoDB
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=notification_service

# RabbitMQ
RABBITMQ_URL=******localhost:5672/

# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-password
SMTP_FROM_EMAIL=noreply@example.com
SMTP_FROM_NAME=Notification Service
SMTP_POOL_SIZE=10
EMAIL_WORKERS=5

# SMS Configuration
SMS_PROVIDER=twilio  # or 'sns'
TWILIO_SID=your-account-sid
TWILIO_TOKEN=your-auth-token
TWILIO_FROM=+1234567890

# AWS Configuration (for SES/SNS)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key

# Rate Limiting
RATE_LIMIT_PER_TENANT=100
RATE_LIMIT_BURST=200

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

See [DEPENDENCIES.md](docs/DEPENDENCIES.md) for a complete list of environment variables.

## Usage

### Send Email Notification

```bash
curl -X POST http://localhost:8084/api/v1/notifications/email \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant123",
    "to": "user@example.com",
    "subject": "Welcome!",
    "body": "Welcome to our service!",
    "template_id": "welcome-email",
    "variables": {
      "user_name": "John Doe",
      "company_name": "Acme Corp"
    }
  }'
```

### Send SMS Notification

```bash
curl -X POST http://localhost:8084/api/v1/notifications/sms \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant123",
    "to": "+14155552671",
    "message": "Your verification code is: 123456"
  }'
```

### Send Webhook Notification

```bash
curl -X POST http://localhost:8084/api/v1/notifications/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant123",
    "url": "https://example.com/webhook",
    "method": "POST",
    "payload": {
      "event": "user.created",
      "data": {"user_id": "123"}
    }
  }'
```

### Bulk Email Sending

```bash
curl -X POST http://localhost:8084/api/v1/notifications/bulk/email \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant123",
    "template_id": "newsletter",
    "recipients": [
      {"email": "user1@example.com", "variables": {"name": "User 1"}},
      {"email": "user2@example.com", "variables": {"name": "User 2"}}
    ]
  }'
```

### Schedule Notification

```bash
curl -X POST http://localhost:8084/api/v1/scheduled \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant123",
    "schedule": "0 9 * * *",
    "notification": {
      "type": "email",
      "to": "user@example.com",
      "subject": "Daily Report",
      "template_id": "daily-report"
    }
  }'
```

### Get Notification History

```bash
curl "http://localhost:8084/api/v1/notifications?tenant_id=tenant123&page=1&page_size=20"
```

### Manage User Preferences

```bash
# Get preferences
curl "http://localhost:8084/api/v1/preferences/user123"

# Update preferences
curl -X PUT http://localhost:8084/api/v1/preferences/user123 \
  -H "Content-Type: application/json" \
  -d '{
    "email_enabled": true,
    "sms_enabled": false,
    "categories": {
      "marketing": false,
      "transactional": true,
      "alerts": true
    }
  }'
```

### Dead Letter Queue Operations

```bash
# Get failed notifications
curl "http://localhost:8084/api/v1/dlq"

# Retry failed notification
curl -X POST "http://localhost:8084/api/v1/dlq/notification123/retry"
```

## Development

### Running Locally

```bash
# Run the service
make run

# Or with go run
go run cmd/main.go
```

### Running with Docker

```bash
# Build and run
make docker-build
make docker-run
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

### Linting

```bash
# Run linters
make lint

# Format code
make fmt
```

## API Documentation

### Endpoints

#### Notifications
- `POST /api/v1/notifications/email` - Send email notification
- `POST /api/v1/notifications/sms` - Send SMS notification
- `POST /api/v1/notifications/webhook` - Send webhook notification
- `GET /api/v1/notifications` - List notifications
- `GET /api/v1/notifications/:id` - Get notification details

#### Bulk Operations
- `POST /api/v1/notifications/bulk/email` - Send bulk emails

#### Scheduled Notifications
- `GET /api/v1/scheduled` - List scheduled notifications
- `POST /api/v1/scheduled` - Create schedule
- `PUT /api/v1/scheduled/:id` - Update schedule
- `DELETE /api/v1/scheduled/:id` - Delete schedule

#### Preferences
- `GET /api/v1/preferences/:user_id` - Get user preferences
- `PUT /api/v1/preferences/:user_id` - Update preferences

#### Dead Letter Queue
- `GET /api/v1/dlq` - List failed notifications
- `POST /api/v1/dlq/:id/retry` - Retry failed notification

#### Webhooks (Provider Callbacks)
- `POST /webhooks/ses` - AWS SES bounce webhook
- `POST /webhooks/sendgrid` - SendGrid bounce webhook

#### Health & Monitoring
- `GET /health` - Service health check
- `GET /ready` - Service readiness check
- `GET /metrics` - Prometheus metrics

See full API documentation at [docs/API.md](docs/API.md).

## Deployment

### Docker

```bash
# Build image
docker build -t notification-service:latest .

# Run container
docker run -d \
  --name notification-service \
  -p 8084:8084 \
  -e MONGODB_URI=mongodb://mongo:27017 \
  -e RABBITMQ_URL=******rabbitmq:5672/ \
  notification-service:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  notification-service:
    build: .
    ports:
      - "8084:8084"
    environment:
      - MONGODB_URI=mongodb://mongo:27017
      - RABBITMQ_URL=******rabbitmq:5672/
    depends_on:
      - mongodb
      - rabbitmq
```

### Kubernetes

```bash
# Apply manifests
kubectl apply -f k8s/

# Check deployment
kubectl get pods -l app=notification-service
```

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed deployment instructions.

## Monitoring

### Prometheus Metrics

The service exposes metrics at `/metrics` endpoint:

```
# Notification metrics
notification_sent_total{channel="email",status="success"}
notification_sent_total{channel="email",status="failed"}
notification_latency_seconds{channel="email",quantile="0.99"}

# SMTP pool metrics
smtp_pool_size
smtp_pool_available
smtp_pool_in_use

# Queue metrics
queue_depth{queue="notification_queue"}
queue_consumer_count

# HTTP metrics
http_requests_total{method="POST",path="/api/v1/notifications/email",status="200"}
http_request_duration_seconds
```

### Grafana Dashboards

Import pre-built dashboards from `monitoring/grafana/` directory.

### Alerting Rules

Example Prometheus alerting rules:

```yaml
groups:
  - name: notification-service
    rules:
      - alert: HighNotificationFailureRate
        expr: rate(notification_sent_total{status="failed"}[5m]) > 0.1
        for: 5m
        annotations:
          summary: "High notification failure rate"
          
      - alert: SMTPPoolExhausted
        expr: smtp_pool_available == 0
        for: 2m
        annotations:
          summary: "SMTP connection pool exhausted"
```

## Architecture

The service follows a clean architecture pattern with clear separation of concerns:

```
├── cmd/                    # Application entry point
├── internal/
│   ├── consumer/          # RabbitMQ event consumers
│   ├── dlq/               # Dead Letter Queue implementation
│   ├── domain/            # Domain models and entities
│   ├── handler/           # HTTP request handlers
│   ├── metrics/           # Prometheus metrics
│   ├── middleware/        # HTTP middleware (rate limiting, etc.)
│   ├── queue/             # Priority queue implementation
│   ├── repository/        # Data access layer
│   ├── scheduler/         # Cron-based scheduler
│   ├── service/           # Business logic
│   ├── shared/            # Shared utilities (logger, config, etc.)
│   ├── smtp/              # SMTP connection pool
│   └── webhook/           # Webhook handlers
└── docs/
    ├── diagrams/          # PlantUML architecture diagrams
    ├── PROVIDER_INTEGRATION.md
    ├── TEMPLATE_BEST_PRACTICES.md
    └── TROUBLESHOOTING.md
```

### Architecture Diagrams

- [Notification Architecture](docs/diagrams/notification-architecture.puml) - Overall system architecture
- [Notification Flow](docs/diagrams/notification-flow.puml) - End-to-end delivery flow
- [Multi-Channel Routing](docs/diagrams/multi-channel-routing.puml) - Channel selection logic
- [Template Processing](docs/diagrams/template-processing.puml) - Template rendering workflow
- [Retry Mechanism](docs/diagrams/retry-mechanism.puml) - Failure retry strategy
- [Queue Architecture](docs/diagrams/queue-architecture.puml) - Message queue design
- [Provider Integration](docs/diagrams/provider-integration.puml) - External provider connections

### Key Design Patterns

- **Strategy Pattern**: Different notification channels (Email, SMS, Push, Webhook)
- **Factory Pattern**: Provider creation based on configuration
- **Observer Pattern**: Delivery tracking and metrics
- **Circuit Breaker**: Provider failure protection
- **Repository Pattern**: Data access abstraction

## Documentation

- [Provider Integration Guide](docs/PROVIDER_INTEGRATION.md) - Configure email, SMS, and push providers
- [Template Best Practices](docs/TEMPLATE_BEST_PRACTICES.md) - Create effective notification templates
- [Troubleshooting Guide](docs/TROUBLESHOOTING.md) - Common issues and solutions
- [Dependencies](docs/DEPENDENCIES.md) - Environment variables and dependencies
- [Architecture Diagrams](docs/diagrams/) - PlantUML diagrams

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run linters: `make lint`
6. Run tests: `make test`
7. Submit a pull request

## Performance

### Benchmarks

- Email throughput: ~1000 emails/second (with connection pooling)
- SMS throughput: ~500 SMS/second
- Webhook throughput: ~2000 requests/second
- Average latency: <100ms (p95)

### Scalability

- Horizontal scaling: Add more instances behind load balancer
- Vertical scaling: Increase worker count and connection pool size
- Database: MongoDB with replica set for high availability
- Queue: RabbitMQ cluster for reliability

## Security

- API key authentication
- Per-tenant rate limiting
- Content validation and sanitization
- XSS protection in templates
- HTTPS only for webhooks
- Encrypted credentials storage
- Audit logging

## Troubleshooting

Common issues and solutions:

- **Emails not sending**: Check SMTP credentials and connection
- **High latency**: Increase connection pool size
- **Messages stuck in queue**: Check consumer status
- **Template errors**: Validate template syntax

See [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for detailed troubleshooting guide.

## Related Repositories

- [go-shared](https://github.com/vhvcorp/go-shared) - Shared Go libraries (deprecated, now using internal/shared)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and changes.

## License

MIT License - see [LICENSE](LICENSE) for details

## Support

- **Documentation**: [GitHub Wiki](https://github.com/vhvcorp/go-notification-service/wiki)
- **Issues**: [GitHub Issues](https://github.com/vhvcorp/go-notification-service/issues)
- **Discussions**: [GitHub Discussions](https://github.com/vhvcorp/go-notification-service/discussions)
- **Email**: support@vhvcorp.com

## Acknowledgments

Built with:
- [Gin](https://github.com/gin-gonic/gin) - HTTP framework
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) - Database driver
- [RabbitMQ](https://www.rabbitmq.com/) - Message broker
- [Prometheus](https://prometheus.io/) - Monitoring and alerting
