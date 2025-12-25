# Notification Service Dependencies

## Shared Packages (from go-shared)

```go
require (
    github.com/vhvcorp/go-shared/config
    github.com/vhvcorp/go-shared/logger
    github.com/vhvcorp/go-shared/mongodb
    github.com/vhvcorp/go-shared/rabbitmq
    github.com/vhvcorp/go-shared/errors
    github.com/vhvcorp/go-shared/middleware
    github.com/vhvcorp/go-shared/response
)
```

## External Dependencies

### Infrastructure
- **MongoDB**: Notification history, templates
  - Collections: `notifications`, `notification_templates`, `notification_logs`
- **RabbitMQ**: Event-driven notifications
  - Queues: `notifications`, `email_queue`, `webhook_queue`
- **SMTP**: Email sending (optional external service)

### Third-party Libraries
```go
require (
    github.com/gin-gonic/gin v1.10.0
    google.golang.org/grpc v1.69.2
    go.mongodb.org/mongo-driver v1.17.3
    github.com/rabbitmq/amqp091-go v1.10.0
    github.com/robfig/cron/v3 v3.0.1
)
```

## Inter-service Communication

### Services Called by Notification Service
- None (leaf service)

### Services Calling Notification Service
- **User Service**: User-related notifications
- **Auth Service**: Authentication notifications

## Environment Variables

```bash
# Server
NOTIFICATION_SERVICE_PORT=50054
NOTIFICATION_SERVICE_HTTP_PORT=8084

# Database
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=saas_framework

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# SMTP
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=your-password
SMTP_FROM=noreply@example.com

# Logging
LOG_LEVEL=info
```

## Database Schema

### Collections

#### notifications
```json
{
  "_id": "ObjectId",
  "user_id": "string (indexed)",
  "type": "string",
  "title": "string",
  "message": "string",
  "status": "string (indexed)",
  "sent_at": "timestamp",
  "created_at": "timestamp"
}
```

## Resource Requirements

### Production
- CPU: 1 core
- Memory: 1GB
- Replicas: 2
