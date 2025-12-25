# Troubleshooting Guide

## Overview

This guide helps you diagnose and resolve common issues with the Notification Service.

## Table of Contents

- [Service Won't Start](#service-wont-start)
- [Email Delivery Issues](#email-delivery-issues)
- [SMS Delivery Issues](#sms-delivery-issues)
- [Push Notification Issues](#push-notification-issues)
- [Webhook Issues](#webhook-issues)
- [Performance Problems](#performance-problems)
- [Database Issues](#database-issues)
- [Queue Issues](#queue-issues)
- [Template Problems](#template-problems)
- [Authentication & Authorization](#authentication--authorization)

## Service Won't Start

### MongoDB Connection Failure

**Symptoms:**
```
Failed to connect to MongoDB: connection refused
```

**Solutions:**
1. Verify MongoDB is running:
   ```bash
   systemctl status mongod
   ```

2. Check connection string:
   ```bash
   echo $MONGODB_URI
   # Should be: mongodb://localhost:27017
   ```

3. Test connectivity:
   ```bash
   mongo --host localhost --port 27017
   ```

4. Check firewall rules:
   ```bash
   sudo firewall-cmd --list-all
   ```

### RabbitMQ Connection Failure

**Symptoms:**
```
Failed to connect to RabbitMQ: dial tcp connection refused
```

**Solutions:**
1. Verify RabbitMQ is running:
   ```bash
   systemctl status rabbitmq-server
   ```

2. Check connection URL:
   ```bash
   echo $RABBITMQ_URL
   ```

3. Verify credentials:
   ```bash
   rabbitmqctl list_users
   ```

4. Check RabbitMQ logs:
   ```bash
   tail -f /var/log/rabbitmq/rabbit@hostname.log
   ```

### Port Already in Use

**Symptoms:**
```
bind: address already in use
```

**Solutions:**
1. Find process using port:
   ```bash
   lsof -i :8084
   ```

2. Kill the process:
   ```bash
   kill -9 <PID>
   ```

3. Or use different port:
   ```bash
   export NOTIFICATION_SERVICE_PORT=8085
   ```

## Email Delivery Issues

### Emails Not Sending

**Check List:**
1. ✅ SMTP credentials correct?
2. ✅ SMTP host reachable?
3. ✅ Sender email verified?
4. ✅ Recipient email valid?
5. ✅ Rate limits not exceeded?

**Debug Steps:**

1. Enable debug logging:
   ```bash
   export LOG_LEVEL=debug
   ```

2. Check SMTP pool status:
   ```bash
   curl http://localhost:8084/api/v1/internal/smtp/pool/status
   ```

3. Test SMTP connection:
   ```bash
   telnet smtp.gmail.com 587
   ```

4. Check notification logs:
   ```bash
   curl http://localhost:8084/api/v1/notifications?status=failed
   ```

### Emails Going to Spam

**Common Causes:**
- Missing SPF record
- Missing DKIM signature
- Missing DMARC policy
- Poor sender reputation
- Spam trigger words

**Solutions:**

1. Verify DNS records:
   ```bash
   dig TXT example.com
   # Look for SPF: v=spf1 include:_spf.google.com ~all
   ```

2. Set up DKIM:
   ```bash
   # Generate DKIM key
   openssl genrsa -out dkim_private.key 1024
   openssl rsa -in dkim_private.key -pubout -out dkim_public.key
   ```

3. Add DMARC record:
   ```
   _dmarc.example.com TXT "v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"
   ```

4. Monitor sender score:
   - Check: https://www.senderscore.org/
   - Check: https://postmaster.google.com/

### High Bounce Rate

**Investigation:**

1. Check bounce repository:
   ```bash
   curl http://localhost:8084/api/v1/bounces?type=hard
   ```

2. Analyze bounce reasons:
   ```sql
   db.bounces.aggregate([
     { $group: { _id: "$reason", count: { $sum: 1 } } },
     { $sort: { count: -1 } }
   ])
   ```

3. Common bounce reasons:
   - **Invalid email**: Remove from list
   - **Mailbox full**: Retry later
   - **Spam block**: Improve content
   - **Domain not found**: Verify domain

**Solutions:**
- Implement email verification
- Clean inactive subscribers
- Monitor bounce webhooks
- Use double opt-in

## SMS Delivery Issues

### SMS Not Delivering

**Check List:**
1. ✅ Valid phone number format?
2. ✅ Country code included?
3. ✅ Provider credentials valid?
4. ✅ Sufficient account balance?
5. ✅ Number not blacklisted?

**Debug Steps:**

1. Verify phone number format:
   ```bash
   # Must include country code
   # Good: +14155552671
   # Bad: 4155552671
   ```

2. Check Twilio logs:
   ```bash
   curl -X GET https://api.twilio.com/2010-04-01/Accounts/{AccountSid}/Messages.json \
     -u {AccountSid}:{AuthToken}
   ```

3. Test with curl:
   ```bash
   curl -X POST http://localhost:8084/api/v1/notifications/sms \
     -H "Content-Type: application/json" \
     -d '{
       "to": "+14155552671",
       "message": "Test SMS",
       "tenant_id": "test"
     }'
   ```

### SMS Delayed

**Possible Causes:**
- Carrier delays
- High volume queue
- Rate limiting
- Geographic routing

**Solutions:**
1. Check queue depth:
   ```bash
   rabbitmqctl list_queues name messages
   ```

2. Increase workers:
   ```bash
   export SMS_WORKERS=10
   ```

3. Monitor provider status:
   - Twilio: https://status.twilio.com/
   - AWS: https://status.aws.amazon.com/

## Push Notification Issues

### Device Not Receiving

**Debug Steps:**

1. Verify device token:
   ```bash
   curl -X POST http://localhost:8084/api/v1/devices/verify \
     -d '{"token": "device-token", "platform": "ios"}'
   ```

2. Check FCM/APNs response:
   - Look for invalid token errors
   - Check for expired certificates

3. Test direct send:
   ```bash
   # FCM
   curl -X POST https://fcm.googleapis.com/fcm/send \
     -H "Authorization: Bearer {token}" \
     -H "Content-Type: application/json" \
     -d '{
       "to": "device-token",
       "notification": {
         "title": "Test",
         "body": "Test notification"
       }
     }'
   ```

### Certificate Expired (APNs)

**Symptoms:**
```
certificate has expired or is not valid yet
```

**Solutions:**
1. Check certificate expiration:
   ```bash
   openssl pkcs8 -in cert.p8 -inform PEM -outform PEM | openssl x509 -noout -dates
   ```

2. Generate new certificate in Apple Developer portal

3. Update service configuration:
   ```bash
   export APNS_CERTIFICATE_PATH=/path/to/new/cert.p8
   ```

## Webhook Issues

### Webhook Not Receiving

**Debug Steps:**

1. Check webhook configuration:
   ```bash
   curl http://localhost:8084/api/v1/webhooks/config
   ```

2. Test webhook endpoint:
   ```bash
   curl -X POST https://your-endpoint.com/webhook \
     -H "Content-Type: application/json" \
     -d '{"test": true}'
   ```

3. Check webhook logs:
   ```bash
   curl http://localhost:8084/api/v1/webhooks/logs?endpoint=https://your-endpoint.com
   ```

4. Verify network connectivity:
   ```bash
   curl -v https://your-endpoint.com
   ```

### Webhook Signature Verification Failing

**Symptoms:**
```
invalid webhook signature
```

**Solutions:**

1. Verify signature algorithm:
   ```go
   // Expected format
   signature := hmac.SHA256(secret, body)
   ```

2. Check secret key:
   ```bash
   echo $WEBHOOK_SECRET
   ```

3. Debug signature:
   ```bash
   # Log incoming signature header
   X-Webhook-Signature: sha256=abc123...
   ```

## Performance Problems

### High Latency

**Investigation:**

1. Check service metrics:
   ```bash
   curl http://localhost:8084/metrics | grep latency
   ```

2. Identify bottleneck:
   ```bash
   # Database queries
   db.system.profile.find({millis: {$gt: 100}})
   
   # SMTP pool
   curl http://localhost:8084/api/v1/internal/smtp/pool/stats
   ```

3. Enable profiling:
   ```bash
   go tool pprof http://localhost:8084/debug/pprof/profile
   ```

**Solutions:**
- Increase connection pool size
- Add database indexes
- Enable template caching
- Scale horizontally

### High Memory Usage

**Investigation:**

1. Check memory stats:
   ```bash
   curl http://localhost:8084/debug/pprof/heap > heap.prof
   go tool pprof -http=:8080 heap.prof
   ```

2. Look for leaks:
   ```bash
   go tool pprof -alloc_space http://localhost:8084/debug/pprof/heap
   ```

**Solutions:**
- Reduce cache size
- Implement connection pooling
- Add memory limits
- Review goroutine leaks

### High CPU Usage

**Investigation:**

1. CPU profiling:
   ```bash
   go tool pprof http://localhost:8084/debug/pprof/profile?seconds=30
   ```

2. Check goroutines:
   ```bash
   curl http://localhost:8084/debug/pprof/goroutine?debug=1
   ```

**Solutions:**
- Optimize template rendering
- Reduce worker count
- Add rate limiting
- Review infinite loops

## Database Issues

### Slow Queries

**Investigation:**

1. Enable profiling:
   ```javascript
   db.setProfilingLevel(2)
   ```

2. Find slow queries:
   ```javascript
   db.system.profile.find({millis: {$gt: 100}}).sort({millis: -1})
   ```

3. Explain query:
   ```javascript
   db.notifications.find({tenant_id: "123"}).explain("executionStats")
   ```

**Solutions:**

1. Add missing indexes:
   ```javascript
   db.notifications.createIndex({tenant_id: 1, created_at: -1})
   db.notifications.createIndex({status: 1, created_at: -1})
   db.templates.createIndex({tenant_id: 1, template_id: 1})
   ```

2. Optimize queries:
   ```javascript
   // Bad: Full collection scan
   db.notifications.find({})
   
   // Good: Indexed query with limit
   db.notifications.find({tenant_id: "123"}).limit(100)
   ```

### Connection Pool Exhausted

**Symptoms:**
```
connection pool timeout
```

**Solutions:**

1. Increase pool size:
   ```bash
   export MONGODB_MAX_POOL_SIZE=100
   ```

2. Reduce connection timeout:
   ```bash
   export MONGODB_CONNECT_TIMEOUT=5s
   ```

3. Check for connection leaks:
   ```bash
   # Monitor active connections
   db.serverStatus().connections
   ```

## Queue Issues

### Messages Not Processing

**Investigation:**

1. Check queue status:
   ```bash
   rabbitmqctl list_queues name messages consumers
   ```

2. Check consumer status:
   ```bash
   curl http://localhost:8084/api/v1/internal/consumers/status
   ```

3. Look for errors:
   ```bash
   rabbitmqctl list_channels
   ```

**Solutions:**

1. Restart consumer:
   ```bash
   curl -X POST http://localhost:8084/api/v1/internal/consumers/restart
   ```

2. Purge queue (caution):
   ```bash
   rabbitmqctl purge_queue notification_queue
   ```

3. Check DLQ:
   ```bash
   curl http://localhost:8084/api/v1/dlq
   ```

### Message Redelivery Loop

**Symptoms:**
- Same message processed repeatedly
- High queue churn
- Errors in logs

**Solutions:**

1. Check message TTL:
   ```bash
   rabbitmqctl list_queues name arguments
   ```

2. Fix application logic
3. Move to DLQ manually:
   ```bash
   curl -X POST http://localhost:8084/api/v1/dlq/move/{messageId}
   ```

## Template Problems

### Variables Not Replaced

**Debug:**

1. Check template syntax:
   ```bash
   curl http://localhost:8084/api/v1/templates/{id}
   ```

2. Verify variable names:
   ```json
   {
     "template": "Hello {{user_name}}",
     "variables": {"user_name": "John"}
   }
   ```

3. Test rendering:
   ```bash
   curl -X POST http://localhost:8084/api/v1/templates/preview \
     -d '{"template_id": "test", "variables": {...}}'
   ```

### Template Not Found

**Solutions:**

1. List available templates:
   ```bash
   curl http://localhost:8084/api/v1/templates
   ```

2. Check cache:
   ```bash
   curl http://localhost:8084/api/v1/internal/cache/templates
   ```

3. Clear cache:
   ```bash
   curl -X POST http://localhost:8084/api/v1/internal/cache/clear
   ```

## Authentication & Authorization

### API Key Invalid

**Solutions:**

1. Verify API key format:
   ```bash
   curl -H "X-API-Key: your-key" http://localhost:8084/api/v1/notifications
   ```

2. Check key expiration:
   ```bash
   curl http://localhost:8084/api/v1/keys/{keyId}
   ```

3. Regenerate key:
   ```bash
   curl -X POST http://localhost:8084/api/v1/keys/regenerate
   ```

### Rate Limit Exceeded

**Symptoms:**
```
429 Too Many Requests
```

**Solutions:**

1. Check current rate limit:
   ```bash
   curl -I http://localhost:8084/api/v1/notifications
   # Look for: X-RateLimit-Limit, X-RateLimit-Remaining
   ```

2. Request limit increase
3. Implement backoff:
   ```go
   if response.StatusCode == 429 {
       retryAfter := response.Header.Get("Retry-After")
       time.Sleep(time.Duration(retryAfter) * time.Second)
   }
   ```

## Getting Help

### Logs

Enable detailed logging:
```bash
export LOG_LEVEL=debug
export LOG_FORMAT=json
```

View logs:
```bash
journalctl -u notification-service -f
```

### Metrics

Access Prometheus metrics:
```bash
curl http://localhost:8084/metrics
```

Key metrics to monitor:
- `notification_sent_total`
- `notification_failed_total`
- `notification_latency_seconds`
- `smtp_pool_size`
- `queue_depth`

### Health Checks

```bash
curl http://localhost:8084/health
curl http://localhost:8084/ready
```

### Support Channels

- GitHub Issues: https://github.com/vhvcorp/go-notification-service/issues
- Documentation: https://github.com/vhvcorp/go-notification-service/wiki
- Email: support@vhvcorp.com

### Reporting Bugs

Include:
1. Service version
2. Error logs
3. Steps to reproduce
4. Configuration (redact secrets)
5. Environment details
