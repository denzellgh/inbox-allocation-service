package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with context-aware methods
type Logger struct {
	*zap.Logger
}

// New creates a new configured logger
func New(level string, format string) (*Logger, error) {
	var config zap.Config

	// Choose configuration based on format
	if format == "console" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Set log level
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	config.Level = zap.NewAtomicLevelAt(zapLevel)

	// Add caller info
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Build logger
	zapLogger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{Logger: zapLogger}, nil
}

// NewFromZap wraps an existing zap.Logger
func NewFromZap(zapLogger *zap.Logger) *Logger {
	return &Logger{Logger: zapLogger}
}

// WithContext creates a child logger with context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}

	fields := make([]zap.Field, 0, 3)

	// Extract correlation ID
	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok && correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}

	// Extract tenant ID
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok && tenantID != "" {
		fields = append(fields, zap.String("tenant_id", tenantID))
	}

	// Extract operator ID
	if operatorID, ok := ctx.Value(OperatorIDKey).(string); ok && operatorID != "" {
		fields = append(fields, zap.String("operator_id", operatorID))
	}

	if len(fields) == 0 {
		return l
	}

	return &Logger{Logger: l.Logger.With(fields...)}
}

// WithFields adds fields to logger
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(fields...)}
}

// WithError adds error field
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Error(err))}
}

// WithCorrelationID adds correlation ID
func (l *Logger) WithCorrelationID(id string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("correlation_id", id))}
}

// WithTenant adds tenant ID
func (l *Logger) WithTenant(tenantID string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("tenant_id", tenantID))}
}

// WithOperator adds operator ID
func (l *Logger) WithOperator(operatorID string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("operator_id", operatorID))}
}

// WithService adds service name
func (l *Logger) WithService(name string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("service", name))}
}

// WithMethod adds method name
func (l *Logger) WithMethod(name string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("method", name))}
}

// Named creates a named child logger
func (l *Logger) Named(name string) *Logger {
	return &Logger{Logger: l.Logger.Named(name)}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() {
	_ = l.Logger.Sync()
}

// Fatal logs a message at FatalLevel and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

// Zap returns the underlying zap.Logger
func (l *Logger) Zap() *zap.Logger {
	return l.Logger
}
