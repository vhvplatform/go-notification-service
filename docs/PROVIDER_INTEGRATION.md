# Provider Integration Guide

## Overview

The Notification Service supports multiple notification channels and providers. This guide explains how to integrate and configure each provider.

## Email Providers

### SMTP Server

Basic email sending using any SMTP-compatible server.

**Configuration:**
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@example.com
SMTP_FROM_NAME=Notification Service
SMTP_POOL_SIZE=10
```

**Features:**
- Connection pooling for performance
- TLS/SSL support
- Authentication support
- Retry on temporary failures

**Best Practices:**
- Use app-specific passwords
- Enable connection pooling for high volume
- Monitor connection pool metrics
- Set appropriate timeout values

### SendGrid

Cloud-based email delivery service with advanced features.

**Configuration:**
```bash
EMAIL_PROVIDER=sendgrid
SENDGRID_API_KEY=your-sendgrid-api-key
SENDGRID_FROM_EMAIL=noreply@example.com
SENDGRID_FROM_NAME=Your Company
```

**Features:**
- High deliverability rates
- Email analytics and tracking
- Template management
- Webhook integration for bounces

**Setup Steps:**
1. Create SendGrid account
2. Generate API key with send permissions
3. Verify sender identity
4. Configure webhook for bounce handling:
   ```
   POST https://your-domain.com/webhooks/sendgrid
   ```

### AWS SES (Simple Email Service)

Amazon's scalable email service.

**Configuration:**
```bash
EMAIL_PROVIDER=ses
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
SES_FROM_EMAIL=noreply@example.com
```

**Features:**
- High throughput
- Cost-effective
- Email reputation dashboard
- SNS integration for notifications

**Setup Steps:**
1. Verify email addresses or domains in SES console
2. Request production access (removal of sandbox)
3. Create IAM user with SES send permissions
4. Configure SNS topic for bounce notifications
5. Set up webhook endpoint for SES notifications:
   ```
   POST https://your-domain.com/webhooks/ses
   ```

## SMS Providers

### Twilio

Leading SMS/voice communication platform.

**Configuration:**
```bash
SMS_PROVIDER=twilio
TWILIO_SID=your-account-sid
TWILIO_TOKEN=your-auth-token
TWILIO_FROM=+1234567890
```

**Features:**
- Global SMS delivery
- Two-way messaging
- Delivery reports
- Number management

**Setup Steps:**
1. Create Twilio account
2. Purchase phone number
3. Get Account SID and Auth Token
4. Configure status callbacks (optional):
   ```
   POST https://your-domain.com/webhooks/twilio/status
   ```

**Rate Limits:**
- Default: 1 message per second
- Can be increased upon request

### AWS SNS (Simple Notification Service)

Amazon's pub/sub and SMS service.

**Configuration:**
```bash
SMS_PROVIDER=sns
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_SNS_ARN=arn:aws:sns:region:account:topic
```

**Features:**
- Multi-channel support (SMS, push, email)
- Topic-based messaging
- Message filtering
- Cost-effective for high volume

**Setup Steps:**
1. Create SNS topic
2. Set up IAM permissions for SMS sending
3. Configure SMS preferences in SNS console
4. Set spending limits

## Push Notifications

### Firebase Cloud Messaging (FCM)

Google's push notification service for mobile apps.

**Configuration:**
```bash
PUSH_PROVIDER=fcm
FCM_SERVER_KEY=your-server-key
FCM_PROJECT_ID=your-project-id
```

**Features:**
- Android and iOS support
- Topic-based messaging
- Device group messaging
- Analytics

**Setup Steps:**
1. Create Firebase project
2. Download service account JSON
3. Enable FCM API
4. Obtain server key from Firebase console

### Apple Push Notification Service (APNs)

Apple's push notification service for iOS devices.

**Configuration:**
```bash
PUSH_PROVIDER=apns
APNS_KEY_ID=your-key-id
APNS_TEAM_ID=your-team-id
APNS_BUNDLE_ID=com.example.app
APNS_CERTIFICATE_PATH=/path/to/cert.p8
APNS_PRODUCTION=true
```

**Features:**
- Silent notifications
- Rich media attachments
- Critical alerts
- Notification grouping

**Setup Steps:**
1. Create Apple Developer account
2. Generate APNs authentication key
3. Download .p8 certificate
4. Configure in Xcode with proper bundle ID

## Webhook Integration

Send notifications to custom HTTP endpoints.

**Configuration:**
```bash
WEBHOOK_TIMEOUT=30s
WEBHOOK_RETRY_COUNT=3
WEBHOOK_RETRY_DELAY=5s
```

**Request Format:**
```json
{
  "notification_id": "unique-id",
  "type": "notification.sent",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "recipient": "user@example.com",
    "subject": "Test Notification",
    "status": "sent"
  }
}
```

**Security:**
- HTTPS only
- HMAC signature verification
- API key authentication
- IP whitelist support

## Multi-Provider Strategy

### Failover Configuration

Automatically switch providers on failure:

```yaml
email:
  primary: sendgrid
  fallback: smtp
  
sms:
  primary: twilio
  fallback: sns
```

### Load Balancing

Distribute load across multiple providers:

```yaml
email:
  providers:
    - sendgrid: 60%  # 60% of traffic
    - ses: 40%       # 40% of traffic
```

## Monitoring and Alerts

### Provider Health Checks

```bash
GET /api/v1/providers/health
```

Response:
```json
{
  "email": {
    "sendgrid": "healthy",
    "ses": "degraded",
    "smtp": "healthy"
  },
  "sms": {
    "twilio": "healthy",
    "sns": "healthy"
  }
}
```

### Metrics to Monitor

1. **Delivery Rate**: Successful sends / Total attempts
2. **Response Time**: Average provider response time
3. **Error Rate**: Failed sends / Total attempts
4. **Queue Depth**: Pending notifications
5. **Cost per Notification**: Track spending by provider

### Alert Thresholds

```yaml
alerts:
  - metric: delivery_rate
    threshold: < 95%
    severity: high
    
  - metric: response_time
    threshold: > 5s
    severity: medium
    
  - metric: error_rate
    threshold: > 5%
    severity: high
```

## Troubleshooting

### Common Issues

**SMTP Connection Failures:**
- Check firewall rules
- Verify credentials
- Test STARTTLS support
- Check rate limits

**SendGrid/SES Bounces:**
- Verify sender domain
- Check SPF/DKIM records
- Monitor reputation score
- Review bounce reasons

**Twilio Errors:**
- Verify phone number format
- Check balance
- Review error codes
- Verify webhook configuration

**Push Notification Issues:**
- Validate device tokens
- Check certificate expiration
- Verify bundle ID
- Test in production mode

### Debugging Tips

1. Enable debug logging:
   ```bash
   LOG_LEVEL=debug
   ```

2. Check provider-specific logs
3. Monitor webhook deliveries
4. Review DLQ for failed messages
5. Use provider dashboards

## Best Practices

1. **Use Connection Pooling**: For SMTP and HTTP-based providers
2. **Implement Circuit Breakers**: Prevent cascading failures
3. **Monitor Bounce Rates**: Keep below 5%
4. **Rotate API Keys**: Regularly update credentials
5. **Test Failover**: Regularly test backup providers
6. **Rate Limit**: Respect provider limits
7. **Cache Templates**: Reduce provider API calls
8. **Batch When Possible**: Use bulk APIs for efficiency

## Cost Optimization

1. **Choose Right Provider**: Match volume with pricing
2. **Use Batch APIs**: Reduce per-message costs
3. **Implement Caching**: Reduce template rendering
4. **Monitor Usage**: Track costs per provider
5. **Optimize Images**: Compress email images
6. **Remove Inactive Users**: Clean up device tokens

## Security Considerations

1. **Encrypt Credentials**: Use secret management
2. **Rotate Keys**: Regular credential rotation
3. **Validate Webhooks**: Verify signatures
4. **Use TLS**: Encrypted connections only
5. **Audit Logs**: Track all provider access
6. **IP Restrictions**: Whitelist provider IPs
7. **Rate Limiting**: Prevent abuse

## Support

For provider-specific issues:
- SendGrid: https://support.sendgrid.com
- AWS SES: https://aws.amazon.com/ses/support
- Twilio: https://support.twilio.com
- FCM: https://firebase.google.com/support
- APNs: https://developer.apple.com/support
