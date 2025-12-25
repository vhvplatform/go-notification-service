# Template Best Practices

## Overview

This guide covers best practices for creating, managing, and using notification templates in the Notification Service.

## Template Structure

### Basic Template

```json
{
  "id": "welcome-email",
  "name": "Welcome Email",
  "type": "email",
  "subject": "Welcome to {{company_name}}!",
  "body": "Hello {{user_name}},\n\nWelcome to our service!",
  "variables": [
    {
      "name": "user_name",
      "type": "string",
      "required": true
    },
    {
      "name": "company_name",
      "type": "string",
      "required": true,
      "default": "Our Company"
    }
  ]
}
```

### HTML Email Template

```json
{
  "id": "order-confirmation",
  "name": "Order Confirmation",
  "type": "email",
  "subject": "Order #{{order_id}} Confirmed",
  "body_html": "<html><body><h1>Order Confirmed</h1><p>Hi {{customer_name}},</p><p>Your order #{{order_id}} has been confirmed.</p><table>{{#items}}<tr><td>{{name}}</td><td>{{quantity}}</td><td>{{price}}</td></tr>{{/items}}</table><p>Total: {{total}}</p></body></html>",
  "body_text": "Order Confirmed\n\nHi {{customer_name}},\n\nYour order #{{order_id}} has been confirmed.\n\nTotal: {{total}}"
}
```

## Variable Types

### Supported Types

1. **String**: Text values
   ```json
   {"name": "user_name", "type": "string"}
   ```

2. **Number**: Numeric values
   ```json
   {"name": "order_total", "type": "number"}
   ```

3. **Boolean**: True/false values
   ```json
   {"name": "is_premium", "type": "boolean"}
   ```

4. **Array**: Lists of items
   ```json
   {"name": "items", "type": "array"}
   ```

5. **Object**: Nested structures
   ```json
   {"name": "user", "type": "object"}
   ```

6. **Date**: Date/time values
   ```json
   {"name": "expiry_date", "type": "date"}
   ```

### Variable Syntax

```
{{variable_name}}           # Simple substitution
{{user.first_name}}         # Nested object
{{items[0].name}}          # Array access
{{amount | currency}}      # With filter
{{#if premium}}...{{/if}}  # Conditional
{{#each items}}...{{/each}} # Loop
```

## Template Features

### Conditional Blocks

```html
{{#if is_premium}}
  <p>Thank you for being a premium member!</p>
  <p>You get {{discount_percent}}% off on all purchases.</p>
{{else}}
  <p>Upgrade to premium for exclusive benefits!</p>
{{/if}}
```

### Loops

```html
<ul>
{{#each items}}
  <li>{{this.name}} - ${{this.price}}</li>
{{/each}}
</ul>
```

### Filters

```
{{name | uppercase}}         # Convert to uppercase
{{name | lowercase}}         # Convert to lowercase
{{name | capitalize}}        # Capitalize first letter
{{date | format_date}}       # Format date
{{amount | currency}}        # Format as currency
{{text | truncate:50}}       # Truncate to 50 chars
{{url | url_encode}}         # URL encode
{{html | escape}}            # HTML escape
```

### Partials

Reusable template components:

```html
{{> header}}
<p>Email content here</p>
{{> footer}}
```

## Design Guidelines

### Email Templates

1. **Use Tables for Layout**: Better email client support
2. **Inline CSS**: Some clients strip `<style>` tags
3. **Alt Text for Images**: Accessibility and fallback
4. **Mobile Responsive**: Use media queries
5. **Plain Text Version**: Always provide alternative
6. **Preheader Text**: First line appears in inbox preview

Example responsive template:

```html
<!DOCTYPE html>
<html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    @media only screen and (max-width: 600px) {
      .container { width: 100% !important; }
      .content { padding: 10px !important; }
    }
  </style>
</head>
<body>
  <table class="container" width="600" cellpadding="0" cellspacing="0">
    <tr>
      <td class="content">
        <!-- Content here -->
      </td>
    </tr>
  </table>
</body>
</html>
```

### SMS Templates

1. **Keep It Short**: 160 characters for single SMS
2. **Clear CTA**: Single, clear call-to-action
3. **No Special Characters**: Avoid emojis unless supported
4. **Include Opt-Out**: "Reply STOP to unsubscribe"
5. **Brand Name**: Include sender name in content

Example:

```
{{company_name}}: Hi {{name}}, your order #{{order_id}} has shipped! Track it here: {{tracking_url}} Reply STOP to opt-out.
```

### Push Notification Templates

1. **Title**: Clear, attention-grabbing (max 50 chars)
2. **Body**: Concise message (max 150 chars)
3. **Action Buttons**: Up to 3 actions
4. **Deep Linking**: Use app-specific URLs
5. **Icon/Badge**: Include branding

Example:

```json
{
  "title": "{{event_name}} starts soon!",
  "body": "Your event begins in {{time_until}}. Don't miss it!",
  "icon": "event_icon",
  "actions": [
    {"label": "View Details", "action": "view_event"},
    {"label": "Set Reminder", "action": "remind_me"}
  ],
  "deep_link": "app://events/{{event_id}}"
}
```

## Security Best Practices

### Input Validation

Always validate and sanitize variables:

```go
// Validate required fields
if user_name == "" {
    return errors.New("user_name is required")
}

// Sanitize HTML
safeHTML := html.EscapeString(userInput)

// Validate email format
if !isValidEmail(email) {
    return errors.New("invalid email format")
}
```

### XSS Prevention

```html
<!-- BAD: Vulnerable to XSS -->
<p>Welcome {{{user_input}}}</p>

<!-- GOOD: Escaped by default -->
<p>Welcome {{user_input}}</p>

<!-- GOOD: Explicitly escaped -->
<p>Welcome {{html_escape user_input}}</p>
```

### Content Filtering

```go
// Block dangerous patterns
blockedPatterns := []string{
    "<script>",
    "javascript:",
    "onerror=",
    "onclick=",
}

for _, pattern := range blockedPatterns {
    if strings.Contains(content, pattern) {
        return errors.New("dangerous content detected")
    }
}
```

## Personalization

### User Segmentation

Create targeted templates for different user groups:

```json
{
  "templates": {
    "new_user": "welcome-new-user",
    "returning_user": "welcome-back",
    "premium_user": "welcome-premium"
  }
}
```

### Dynamic Content

```html
{{#if user.lifetime_value > 1000}}
  <p>As a valued customer, enjoy 20% off your next purchase!</p>
{{else if user.lifetime_value > 500}}
  <p>Thank you for your continued support! Here's 10% off.</p>
{{else}}
  <p>Welcome! Here's a 5% discount on your next order.</p>
{{/if}}
```

### Localization

```json
{
  "template_id": "welcome-email",
  "translations": {
    "en": {
      "subject": "Welcome to {{company_name}}!",
      "greeting": "Hello"
    },
    "es": {
      "subject": "¡Bienvenido a {{company_name}}!",
      "greeting": "Hola"
    },
    "fr": {
      "subject": "Bienvenue à {{company_name}}!",
      "greeting": "Bonjour"
    }
  }
}
```

## Performance Optimization

### Template Caching

Templates are automatically cached:
- Cache TTL: 1 hour
- LRU eviction policy
- Per-tenant isolation

```go
// Cache key: tenant_id:template_id:version
cacheKey := fmt.Sprintf("%s:%s:%s", tenantID, templateID, version)
```

### Lazy Loading

Load templates only when needed:

```go
// Don't pre-load all templates
// Load on-demand with caching
template := templateCache.Get(templateID)
if template == nil {
    template = templateRepo.FindByID(templateID)
    templateCache.Set(templateID, template)
}
```

### Batch Rendering

Process multiple templates in parallel:

```go
// Render templates concurrently
results := make(chan RenderedTemplate, len(notifications))
for _, notif := range notifications {
    go func(n Notification) {
        rendered := templateEngine.Render(n.TemplateID, n.Variables)
        results <- rendered
    }(notif)
}
```

## Testing Templates

### Unit Tests

```go
func TestWelcomeTemplate(t *testing.T) {
    template := loadTemplate("welcome-email")
    variables := map[string]interface{}{
        "user_name": "John Doe",
        "company_name": "Acme Corp",
    }
    
    result := template.Render(variables)
    
    assert.Contains(t, result, "John Doe")
    assert.Contains(t, result, "Acme Corp")
    assert.NotContains(t, result, "{{")
}
```

### Preview Testing

```bash
# Test template rendering
curl -X POST http://localhost:8084/api/v1/templates/preview \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "welcome-email",
    "variables": {
      "user_name": "Test User",
      "company_name": "Test Company"
    }
  }'
```

### A/B Testing

```json
{
  "experiment": "welcome_email_test",
  "variants": [
    {
      "name": "variant_a",
      "template_id": "welcome-v1",
      "traffic": 50
    },
    {
      "name": "variant_b",
      "template_id": "welcome-v2",
      "traffic": 50
    }
  ]
}
```

## Version Control

### Template Versioning

```json
{
  "template_id": "welcome-email",
  "version": "1.2.0",
  "changelog": "Added mobile responsive design",
  "created_at": "2024-01-01T00:00:00Z",
  "status": "active"
}
```

### Migration Strategy

```bash
# Deploy new version
POST /api/v1/templates
{
  "id": "welcome-email",
  "version": "2.0.0",
  "...": "..."
}

# Test with small percentage
PATCH /api/v1/templates/welcome-email/rollout
{
  "version": "2.0.0",
  "percentage": 10
}

# Full rollout
PATCH /api/v1/templates/welcome-email/rollout
{
  "version": "2.0.0",
  "percentage": 100
}

# Rollback if needed
POST /api/v1/templates/welcome-email/rollback
{
  "to_version": "1.2.0"
}
```

## Common Pitfalls

1. **Missing Variables**: Always provide defaults
2. **Large Images**: Optimize before embedding
3. **Too Much HTML**: Keep it simple
4. **No Plain Text**: Always include alternative
5. **Broken Links**: Test all URLs
6. **Hard-coded Values**: Use variables instead
7. **No Preview**: Always test before deploying
8. **Ignoring Mobile**: Test on mobile devices

## Resources

- [Email on Acid](https://www.emailonacid.com/) - Email testing
- [Litmus](https://www.litmus.com/) - Email previews
- [Really Good Emails](https://reallygoodemails.com/) - Design inspiration
- [Can I Email](https://www.caniemail.com/) - HTML/CSS support
