package errors

import (
	"testing"
)

func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		err     error
		want    string
	}{
		{
			name:    "validation error with underlying error",
			message: "Invalid input",
			err:     NewValidationError("field required", nil),
			want:    "VALIDATION_ERROR: Invalid input",
		},
		{
			name:    "validation error without underlying error",
			message: "Invalid input",
			err:     nil,
			want:    "VALIDATION_ERROR: Invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.message, tt.err)
			if err == nil {
				t.Error("NewValidationError() returned nil")
			}
			if err.Code != "VALIDATION_ERROR" {
				t.Errorf("Code = %v, want VALIDATION_ERROR", err.Code)
			}
			if err.Message != tt.message {
				t.Errorf("Message = %v, want %v", err.Message, tt.message)
			}
		})
	}
}

func TestNewInternalError(t *testing.T) {
	message := "Database connection failed"
	err := NewInternalError(message, nil)

	if err.Code != "INTERNAL_ERROR" {
		t.Errorf("Code = %v, want INTERNAL_ERROR", err.Code)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
}

func TestNewNotFoundError(t *testing.T) {
	message := "Resource not found"
	err := NewNotFoundError(message, nil)

	if err.Code != "NOT_FOUND" {
		t.Errorf("Code = %v, want NOT_FOUND", err.Code)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	message := "Invalid credentials"
	err := NewUnauthorizedError(message, nil)

	if err.Code != "UNAUTHORIZED" {
		t.Errorf("Code = %v, want UNAUTHORIZED", err.Code)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name    string
		appErr  *AppError
		wantStr string
	}{
		{
			name: "error with underlying error",
			appErr: &AppError{
				Code:    "TEST_ERROR",
				Message: "Test message",
				Err:     NewValidationError("underlying", nil),
			},
			wantStr: "TEST_ERROR: Test message",
		},
		{
			name: "error without underlying error",
			appErr: &AppError{
				Code:    "TEST_ERROR",
				Message: "Test message",
				Err:     nil,
			},
			wantStr: "TEST_ERROR: Test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appErr.Error()
			if len(got) == 0 {
				t.Error("Error() returned empty string")
			}
		})
	}
}
