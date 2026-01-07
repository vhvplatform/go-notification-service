package logger

import (
	"log"
	"os"
)

// Logger provides a simple logging interface
type Logger struct {
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// Info logs an informational message
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Printf("[INFO] %s %v", msg, keysAndValues)
}

// Error logs an error message
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Printf("[ERROR] %s %v", msg, keysAndValues)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Printf("[DEBUG] %s %v", msg, keysAndValues)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Printf("[WARN] %s %v", msg, keysAndValues)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.logger.Fatalf("[FATAL] %s %v", msg, keysAndValues)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return nil
}
