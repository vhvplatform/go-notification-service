# Notification Service Dependencies

## Shared Packages (from go-shared)

```go
require (
    github.com/vhvplatform/go-shared/config
    github.com/vhvplatform/go-shared/logger
    github.com/vhvplatform/go-shared/mongodb
    github.com/vhvplatform/go-shared/rabbitmq
    github.com/vhvplatform/go-shared/errors
    github.com/vhvplatform/go-shared/middleware
    github.com/vhvplatform/go-shared/response
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

The application loads environment variables from a `.env` file (if present) or from system environment variables. 

### Configuration Setup

1. **For Local Development**:
```bash
# Copy the example file
cp .env.example .env

# Edit with your configuration
nano .env
```

2. **For Docker Deployment**:
```bash
# Option 1: Mount .env file
docker run -v $(pwd)/.env:/root/.env notification-service

# Option 2: Pass environment variables directly
docker run -e MONGODB_URI=mongodb://mongo:27017 \
           -e RABBITMQ_URL=amqp://rabbitmq:5672/ \
           notification-service
```

3. **For Kubernetes**:
```yaml
# Use ConfigMap and Secrets
apiVersion: v1
kind: ConfigMap
metadata:
  name: notification-config
data:
  NOTIFICATION_SERVICE_PORT: "8084"
  MONGODB_URI: "mongodb://mongo:27017"
  # ... other non-sensitive values
---
apiVersion: v1
kind: Secret
metadata:
  name: notification-secrets
type: Opaque
stringData:
  SMTP_PASSWORD: "your-password"
  TWILIO_TOKEN: "your-token"
  # ... other sensitive values
```

### Complete Environment Variable Reference

#### Server Configuration
```bash
# HTTP server port
NOTIFICATION_SERVICE_PORT=8084

#### Database Configuration
```bash
# MongoDB connection string
# Format: mongodb://[username:password@]host[:port]/[defaultAuthDB]
MONGODB_URI=mongodb://localhost:27017

# Database name
MONGODB_DATABASE=notification_service
```

#### Message Queue Configuration
```bash
# RabbitMQ connection URL
# Format: amqp://[username:password@]host[:port]/[vhost]
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

#### SMTP Configuration
```bash
# SMTP server hostname
SMTP_HOST=smtp.gmail.com

# SMTP server port (usually 587 for TLS, 465 for SSL, 25 for unencrypted)
SMTP_PORT=587

# SMTP authentication username
SMTP_USERNAME=your-email@example.com

# SMTP authentication password
SMTP_PASSWORD=your-password

# Sender email address
SMTP_FROM_EMAIL=noreply@example.com

# Sender display name
SMTP_FROM_NAME=Notification Service

# SMTP connection pool size (default: 10)
SMTP_POOL_SIZE=10
```

#### Email Processing Configuration
```bash
# Number of concurrent workers for processing bulk emails (default: 5)
EMAIL_WORKERS=5
```

#### SMS Configuration
```bash
# SMS provider: 'twilio' or 'sns'
SMS_PROVIDER=twilio

# Twilio Configuration (when SMS_PROVIDER=twilio)
TWILIO_SID=your-twilio-account-sid
TWILIO_TOKEN=your-twilio-auth-token
TWILIO_FROM=+1234567890

# AWS SNS Configuration (when SMS_PROVIDER=sns)
AWS_SNS_ARN=arn:aws:sns:us-east-1:123456789012:your-topic
AWS_REGION=us-east-1
```

#### Rate Limiting Configuration
```bash
# Maximum requests per second per tenant (default: 100)
RATE_LIMIT_PER_TENANT=100

# Maximum burst size for rate limiting (default: 200)
RATE_LIMIT_BURST=200
```

### Configuration Precedence

Environment variables are loaded in the following order (highest to lowest priority):
1. **System environment variables** - Set at the OS level
2. **`.env` file** - Loaded from the application root directory
3. **Default values** - Hardcoded defaults in the application

### Security Best Practices

⚠️ **Important Security Notes**:
- **Never commit `.env` files** to version control
- The `.env` file is already in `.gitignore` to prevent accidental commits
- Use `.env.example` as a template - it contains no sensitive data
- In production, prefer system environment variables or secret management systems (AWS Secrets Manager, HashiCorp Vault, Kubernetes Secrets)
- Rotate credentials regularly
- Use different credentials for different environments (dev, staging, production)

### Default Values

If an environment variable is not set, the application uses these defaults:

| Variable | Default Value |
|----------|--------------|
| NOTIFICATION_SERVICE_PORT | 8084 |
| MONGODB_URI | mongodb://localhost:27017 |
| MONGODB_DATABASE | notification_service |
| RABBITMQ_URL | amqp://guest:guest@localhost:5672/ |
| SMTP_HOST | smtp.gmail.com |
| SMTP_PORT | 587 |
| SMTP_FROM_EMAIL | noreply@example.com |
| SMTP_FROM_NAME | Notification Service |
| SMTP_POOL_SIZE | 10 |
| EMAIL_WORKERS | 5 |
| SMS_PROVIDER | twilio |
| RATE_LIMIT_PER_TENANT | 100 |
| RATE_LIMIT_BURST | 200 |

### Troubleshooting

**Problem**: Application not loading `.env` file
- **Solution**: Ensure `.env` file is in the same directory as the running application
- **Note**: The application will NOT fail if `.env` is missing - it will use system environment variables

**Problem**: Environment variables not being recognized
- **Solution**: Check variable names for typos and ensure no extra spaces
- **Solution**: Verify the `.env` file format (KEY=value, no spaces around =)

**Problem**: Sensitive data exposed in logs
- **Solution**: Never log environment variable values
- **Solution**: Review application logs to ensure passwords/tokens are not printed
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
