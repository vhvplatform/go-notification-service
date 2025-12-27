# Security Improvements

## Overview
This document describes the security enhancements implemented in the Go Notification Service to protect against common vulnerabilities and attacks.

## Security Features Implemented

### 1. MongoDB Connection Security

#### URI Validation
**Location**: `internal/shared/mongodb/mongodb.go`

**Protections**:
- Validates MongoDB URI format to prevent injection attacks
- Enforces correct URI scheme (mongodb:// or mongodb+srv://)
- Validates that host is present in URI
- Prevents empty or malformed URIs

**Code Example**:
```go
func validateMongoURI(uri string) error {
    if uri == "" {
        return errors.New("mongodb URI cannot be empty")
    }
    
    parsedURI, err := url.Parse(uri)
    if err != nil {
        return fmt.Errorf("invalid mongodb URI format: %w", err)
    }
    
    if parsedURI.Scheme != "mongodb" && parsedURI.Scheme != "mongodb+srv" {
        return fmt.Errorf("invalid mongodb URI scheme: %s", parsedURI.Scheme)
    }
    
    return nil
}
```

#### Database Name Validation
**Protections**:
- Prevents empty database names
- Validates against path traversal characters (/, \, .)
- Blocks special characters that could cause injection ($, *, <, >, :, |, ?, etc.)

**Security Impact**: Prevents NoSQL injection and path traversal attacks

#### TLS/SSL Enforcement
**Protections**:
- Automatically enables TLS for mongodb+srv:// URIs
- Enforces minimum TLS 1.2
- Validates server certificates

**Code Example**:
```go
if strings.Contains(uri, "mongodb+srv://") || strings.Contains(uri, "tls=true") {
    tlsConfig := &tls.Config{
        MinVersion: tls.VersionTLS12, // Enforce minimum TLS 1.2
    }
    clientOptions.SetTLSConfig(tlsConfig)
}
```

#### Connection Reliability
**Protections**:
- Enables retryable writes and reads
- Uses read preference validation
- Implements proper timeouts

### 2. Template Cache Security

#### Size Limits
**Location**: `internal/repository/template_repository.go`

**Constants**:
```go
const (
    maxCacheSize     = 1000              // Maximum number of cached templates
    maxCacheKeyLen   = 512               // Maximum length of cache key
    maxTemplateSize  = 1024 * 1024       // Maximum template size: 1MB
)
```

**Protections**:
- Prevents memory exhaustion attacks by limiting cache size
- Limits individual template size to prevent DoS
- Implements LRU eviction when cache is full
- Validates cache key length

#### Cache Key Validation
**Protections**:
- Validates cache key is not empty
- Enforces maximum key length (512 bytes)
- Prevents null bytes in keys
- Blocks newline and carriage return characters

**Code Example**:
```go
func validateCacheKey(key string) error {
    if len(key) == 0 {
        return errors.New("cache key cannot be empty")
    }
    if len(key) > maxCacheKeyLen {
        return errors.New("cache key exceeds maximum length")
    }
    if strings.ContainsAny(key, "\x00\n\r") {
        return errors.New("cache key contains invalid characters")
    }
    return nil
}
```

#### Automatic Eviction
**Protections**:
- Implements LRU (Least Recently Used) eviction
- Prevents unbounded memory growth
- Thread-safe eviction with mutex protection

### 3. Email Service Input Validation

#### Security Constants
**Location**: `internal/service/email_service.go`

**Constants**:
```go
const (
    maxEmailLength     = 320  // Maximum email address length per RFC 5321
    maxSubjectLength   = 998  // Maximum subject line length per RFC 5322
    maxBodyLength      = 10 * 1024 * 1024 // Maximum email body size: 10MB
    maxRecipientsCount = 1000 // Maximum recipients per email
    maxVariableKeyLen  = 256  // Maximum variable key length
    maxVariableValLen  = 65536 // Maximum variable value length: 64KB
)
```

#### Input Validation
**Protections**:
- Validates recipient count (min 1, max 1000)
- Enforces subject line length limits
- Limits email body size to prevent DoS
- Validates UTF-8 encoding in all text fields
- Prevents null bytes in variables
- Limits variable key and value sizes

**Code Example**:
```go
func validateEmailInput(to []string, subject, body string, variables map[string]string) error {
    if len(to) == 0 {
        return errors.New("at least one recipient is required")
    }
    if len(to) > maxRecipientsCount {
        return fmt.Errorf("too many recipients: %d (max: %d)", len(to), maxRecipientsCount)
    }
    
    if len(subject) > maxSubjectLength {
        return fmt.Errorf("subject too long: %d bytes (max: %d)", len(subject), maxSubjectLength)
    }
    
    if !utf8.ValidString(subject) {
        return errors.New("subject contains invalid UTF-8 characters")
    }
    
    // ... more validations
}
```

#### Email Address Validation
**Enhanced Protections**:
- Validates email length (max 320 characters per RFC 5321)
- Checks for null bytes, carriage returns, newlines
- Validates UTF-8 encoding
- Uses regex for format validation

**Code Example**:
```go
func (s *EmailService) isValidEmail(email string) bool {
    if len(email) == 0 || len(email) > maxEmailLength {
        return false
    }
    
    if strings.ContainsAny(email, "\x00\r\n") {
        return false
    }
    
    if !utf8.ValidString(email) {
        return false
    }
    
    return s.emailRegex.MatchString(email)
}
```

### 4. XSS Protection (Existing, Enhanced)

**Location**: `internal/service/email_service.go`

**Protection**:
- All template variables are HTML-escaped before replacement
- Uses `html.EscapeString()` to prevent XSS attacks
- Applied in the `applyVariables()` function

**Code Example**:
```go
for key, value := range variables {
    escapedValue := html.EscapeString(value) // Prevent XSS
    placeholder := fmt.Sprintf("{{%s}}", key)
    replacements = append(replacements, placeholder, escapedValue)
}
```

## Security Testing

### Test Coverage
- 18 security-focused unit tests added
- Tests for MongoDB URI validation
- Tests for database name injection prevention
- Tests for cache key validation
- Tests for template size limits
- Tests for email input validation
- Tests for UTF-8 validation
- Tests for null byte prevention

### Test Files
1. `internal/shared/mongodb/mongodb_test.go` - Database security tests
2. `internal/service/email_service_security_test.go` - Email service security tests
3. `internal/repository/template_repository_test.go` - Cache security tests

### Running Security Tests
```bash
# Run all security tests
make test

# Run specific security test suites
go test -v ./internal/shared/mongodb/... -run Security
go test -v ./internal/service/... -run Security
go test -v ./internal/repository/... -run Security
```

## Vulnerability Mitigations

### 1. NoSQL Injection Prevention
**Threat**: Attackers could manipulate database queries through malicious input
**Mitigation**: 
- URI validation prevents malformed connection strings
- Database name validation blocks special characters
- All queries use parameterized BSON objects

### 2. Denial of Service (DoS) Prevention
**Threat**: Resource exhaustion through large payloads or excessive requests
**Mitigation**:
- Cache size limits (max 1000 entries)
- Template size limits (1MB per template)
- Email body size limits (10MB)
- Recipient count limits (1000 per email)
- Variable size limits (64KB per variable)

### 3. Memory Exhaustion Prevention
**Threat**: Unbounded memory growth leading to crashes
**Mitigation**:
- LRU cache eviction
- Template size validation
- Cache key length limits
- Automatic cleanup of expired entries

### 4. Cross-Site Scripting (XSS) Prevention
**Threat**: Malicious scripts in email templates
**Mitigation**:
- HTML escaping of all template variables
- UTF-8 validation
- Null byte prevention

### 5. Header Injection Prevention
**Threat**: CRLF injection in email headers
**Mitigation**:
- Validates email addresses for control characters
- Blocks newlines and carriage returns
- Subject line length limits

### 6. Path Traversal Prevention
**Threat**: Access to unauthorized files/directories
**Mitigation**:
- Database name validation
- Cache key validation
- Blocks directory separators and special characters

## Security Best Practices

### 1. Input Validation
- **Always validate** all external input before processing
- **Fail securely** by rejecting invalid input
- **Log security events** for monitoring and auditing

### 2. Resource Limits
- **Set explicit limits** on all resource-consuming operations
- **Implement timeouts** to prevent hanging operations
- **Use bounded data structures** (e.g., limited cache size)

### 3. Encoding and Escaping
- **Escape HTML** in user-provided content
- **Validate UTF-8** encoding to prevent encoding attacks
- **Block control characters** (null bytes, CRLF, etc.)

### 4. Secure Communications
- **Use TLS 1.2+** for all network connections
- **Validate certificates** when connecting to databases
- **Enable retryable operations** for better reliability

### 5. Defense in Depth
- **Multiple validation layers** at different points
- **Fail-safe defaults** (e.g., reject on validation error)
- **Comprehensive error handling** without exposing internals

## Monitoring and Alerting

### Security Events to Monitor
1. **Failed validation attempts** - May indicate attack attempts
2. **Cache size approaching limit** - Potential DoS attack
3. **Large payload rejections** - DoS prevention triggers
4. **Invalid UTF-8 submissions** - Encoding attack attempts
5. **TLS handshake failures** - MITM attack attempts

### Logging Recommendations
```go
// Log security-relevant events
log.Warn("Email input validation failed", "error", err, "tenant_id", tenantID)
log.Warn("Cache key validation failed", "key_length", len(key))
log.Error("MongoDB URI validation failed", "error", err)
```

## Compliance

### Standards Addressed
- **RFC 5321**: SMTP email address length limits
- **RFC 5322**: Email header length limits
- **OWASP Top 10**: Injection, broken access control, security misconfiguration
- **CWE-89**: NoSQL Injection prevention
- **CWE-79**: XSS prevention
- **CWE-400**: Resource exhaustion prevention

## Future Security Enhancements

### Short-term (Next Sprint)
1. Add rate limiting per IP address
2. Implement API authentication
3. Add audit logging for security events
4. Implement request signing for webhooks

### Medium-term (Next Quarter)
1. Add encryption at rest for sensitive data
2. Implement role-based access control (RBAC)
3. Add security headers (CSP, HSTS, etc.)
4. Implement secrets management integration

### Long-term (Next 6 Months)
1. Security scanning in CI/CD pipeline
2. Penetration testing
3. Security code reviews
4. Compliance certifications (SOC 2, ISO 27001)

## References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [MongoDB Security Checklist](https://www.mongodb.com/docs/manual/administration/security-checklist/)
- [CWE Common Weakness Enumeration](https://cwe.mitre.org/)
- [RFC 5321 - SMTP](https://tools.ietf.org/html/rfc5321)
- [RFC 5322 - Email Format](https://tools.ietf.org/html/rfc5322)
