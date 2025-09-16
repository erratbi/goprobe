package probe

import (
	"context"
	"log"
	"os"
	"time"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger interface for structured logging
type Logger interface {
	Debug(ctx context.Context, msg string, fields map[string]interface{})
	Info(ctx context.Context, msg string, fields map[string]interface{})
	Warn(ctx context.Context, msg string, fields map[string]interface{})
	Error(ctx context.Context, msg string, fields map[string]interface{})
}

// DefaultLogger is a simple implementation of Logger interface
type DefaultLogger struct {
	level  LogLevel
	logger *log.Logger
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger(level LogLevel) *DefaultLogger {
	return &DefaultLogger{
		level:  level,
		logger: log.New(os.Stderr, "[goprobe] ", log.LstdFlags),
	}
}

func (l *DefaultLogger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	if l.level <= LogLevelDebug {
		l.logWithFields("DEBUG", msg, fields)
	}
}

func (l *DefaultLogger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	if l.level <= LogLevelInfo {
		l.logWithFields("INFO", msg, fields)
	}
}

func (l *DefaultLogger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	if l.level <= LogLevelWarn {
		l.logWithFields("WARN", msg, fields)
	}
}

func (l *DefaultLogger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	if l.level <= LogLevelError {
		l.logWithFields("ERROR", msg, fields)
	}
}

func (l *DefaultLogger) logWithFields(level, msg string, fields map[string]interface{}) {
	logMsg := level + " " + msg
	if len(fields) > 0 {
		logMsg += " "
		for k, v := range fields {
			logMsg += k + "=" + toString(v) + " "
		}
	}
	l.logger.Println(logMsg)
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return string(rune(val))
	case time.Duration:
		return val.String()
	default:
		return "unknown"
	}
}

// NoopLogger is a logger that does nothing (for production when logging is disabled)
type NoopLogger struct{}

func (NoopLogger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {}
func (NoopLogger) Info(ctx context.Context, msg string, fields map[string]interface{})  {}
func (NoopLogger) Warn(ctx context.Context, msg string, fields map[string]interface{})  {}
func (NoopLogger) Error(ctx context.Context, msg string, fields map[string]interface{}) {}

// Global logger instance
var globalLogger Logger = &NoopLogger{}

// SetLogger sets the global logger
func SetLogger(logger Logger) {
	globalLogger = logger
}

// GetLogger returns the current global logger
func GetLogger() Logger {
	return globalLogger
}

// Helper functions for logging
func logDebug(ctx context.Context, msg string, fields map[string]interface{}) {
	globalLogger.Debug(ctx, msg, fields)
}

func logInfo(ctx context.Context, msg string, fields map[string]interface{}) {
	globalLogger.Info(ctx, msg, fields)
}

func logWarn(ctx context.Context, msg string, fields map[string]interface{}) {
	globalLogger.Warn(ctx, msg, fields)
}

func logError(ctx context.Context, msg string, fields map[string]interface{}) {
	globalLogger.Error(ctx, msg, fields)
}