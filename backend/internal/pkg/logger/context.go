package logger

import (
	"context"

	"go.uber.org/zap"
)

// ContextKey for storing values in context
type ContextKey string

const (
	// LoggerKey is the key for storing logger in context
	LoggerKey ContextKey = "logger"
	// CorrelationIDKey is the key for correlation ID
	CorrelationIDKey ContextKey = "correlation_id"
	// TenantIDKey is the key for tenant ID
	TenantIDKey ContextKey = "tenant_id"
	// OperatorIDKey is the key for operator ID
	OperatorIDKey ContextKey = "operator_id"
	// RequestIDKey is the key for request ID
	RequestIDKey ContextKey = "request_id"
)

// FromContext extracts logger from context or returns a no-op logger
func FromContext(ctx context.Context) *Logger {
	if ctx == nil {
		return NewNop()
	}

	if l, ok := ctx.Value(LoggerKey).(*Logger); ok && l != nil {
		return l
	}

	return NewNop()
}

// WithLogger adds logger to context
func WithLogger(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, l)
}

// WithCorrelationIDCtx adds correlation ID to context
func WithCorrelationIDCtx(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, id)
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

// WithTenantIDCtx adds tenant ID to context
func WithTenantIDCtx(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, TenantIDKey, id)
}

// GetTenantID extracts tenant ID from context
func GetTenantID(ctx context.Context) string {
	if id, ok := ctx.Value(TenantIDKey).(string); ok {
		return id
	}
	return ""
}

// WithOperatorIDCtx adds operator ID to context
func WithOperatorIDCtx(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, OperatorIDKey, id)
}

// GetOperatorID extracts operator ID from context
func GetOperatorID(ctx context.Context) string {
	if id, ok := ctx.Value(OperatorIDKey).(string); ok {
		return id
	}
	return ""
}

// NewNop creates a no-op logger for testing or when context has no logger
func NewNop() *Logger {
	return &Logger{Logger: zap.NewNop()}
}
