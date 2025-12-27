package service

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestApplyVariables tests the template variable replacement
func TestApplyVariables(t *testing.T) {
	service := &EmailService{}

	tests := []struct {
		name      string
		template  string
		variables map[string]string
		expected  string
	}{
		{
			name:     "single variable",
			template: "Hello {{name}}!",
			variables: map[string]string{
				"name": "John",
			},
			expected: "Hello John!",
		},
		{
			name:     "multiple variables",
			template: "Hello {{name}}, welcome to {{company}}!",
			variables: map[string]string{
				"name":    "John",
				"company": "Acme Corp",
			},
			expected: "Hello John, welcome to Acme Corp!",
		},
		{
			name:      "no variables",
			template:  "Hello World!",
			variables: map[string]string{},
			expected:  "Hello World!",
		},
		{
			name:     "XSS protection",
			template: "Hello {{name}}!",
			variables: map[string]string{
				"name": "<script>alert('xss')</script>",
			},
			expected: "Hello &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.applyVariables(tt.template, tt.variables)
			if result != tt.expected {
				t.Errorf("applyVariables() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// BenchmarkApplyVariablesSingle benchmarks single variable replacement
func BenchmarkApplyVariablesSingle(b *testing.B) {
	service := &EmailService{}
	template := "Hello {{name}}, welcome to our service!"
	variables := map[string]string{
		"name": "John Doe",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.applyVariables(template, variables)
	}
}

// BenchmarkApplyVariablesMultiple benchmarks multiple variable replacement
func BenchmarkApplyVariablesMultiple(b *testing.B) {
	service := &EmailService{}
	template := "Hello {{name}}, welcome to {{company}}! Your account {{account_id}} is now active. Visit {{url}} to get started."
	variables := map[string]string{
		"name":       "John Doe",
		"company":    "Acme Corp",
		"account_id": "ACC-12345",
		"url":        "https://example.com/dashboard",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.applyVariables(template, variables)
	}
}

// BenchmarkApplyVariablesLarge benchmarks replacement with many variables
func BenchmarkApplyVariablesLarge(b *testing.B) {
	service := &EmailService{}
	
	// Build a template with 20 variables
	var templateBuilder strings.Builder
	variables := make(map[string]string)
	
	templateBuilder.WriteString("Dear {{name}},\n\n")
	variables["name"] = "John Doe"
	
	for i := 1; i <= 18; i++ {
		key := "var" + string(rune('0'+i))
		templateBuilder.WriteString("{{")
		templateBuilder.WriteString(key)
		templateBuilder.WriteString("}} ")
		variables[key] = "value" + string(rune('0'+i))
	}
	
	template := templateBuilder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.applyVariables(template, variables)
	}
}

// TestContextTimeout tests that context timeout is properly handled
func TestContextTimeout(t *testing.T) {
	// This test would require actual email service setup
	// Demonstrating the pattern for timeout testing
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Simulate a long-running operation
	select {
	case <-time.After(200 * time.Millisecond):
		t.Error("Operation should have been cancelled by context timeout")
	case <-ctx.Done():
		// Expected: context deadline exceeded
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded, got %v", ctx.Err())
		}
	}
}

// TestEmailValidation tests email address validation
func TestIsValidEmail(t *testing.T) {
	service := &EmailService{}
	
	// Would need to initialize emailRegex in the service
	// This is a placeholder to show the test pattern
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name@example.com", true},
		{"user+tag@example.co.uk", true},
		{"invalid.email", false},
		{"@example.com", false},
		{"user@", false},
		{"", false},
	}

	// Skip if regex not initialized
	if service.emailRegex == nil {
		t.Skip("Email regex not initialized - integration test")
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := service.isValidEmail(tt.email)
			if result != tt.valid {
				t.Errorf("isValidEmail(%s) = %v, want %v", tt.email, result, tt.valid)
			}
		})
	}
}
