package utils

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

// InitLogger initializes the global logger
func InitLogger(environment, level string) error {
	var cfg zap.Config

	if environment == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	// Set log level
	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	var err error
	logger, err = cfg.Build()
	if err != nil {
		return err
	}

	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if logger == nil {
		// Initialize with default config if not already initialized
		_ = InitLogger("development", "info")
	}
	return logger
}

// Sync flushes any buffered log entries
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}
