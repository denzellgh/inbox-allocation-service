package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// New creates a new configured zap logger
func New(level string, format string) (*zap.Logger, error) {
	var config zap.Config

	// Choose configuration based on format
	if format == "console" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	// Set log level
	var zapLevel zap.AtomicLevel
	switch level {
	case "debug":
		zapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	config.Level = zapLevel

	// Build logger
	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}
