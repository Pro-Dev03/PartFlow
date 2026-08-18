package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Logger is the global logger instance
	Logger zerolog.Logger
)

// Config holds logger configuration
type Config struct {
	Level          string
	OutputPath     string
	EnableConsole  bool
	EnableFile     bool
	TimeFormat     string
	EnableCaller   bool
	EnableStackTrace bool
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:          "info",
		OutputPath:     "logs/app.log",
		EnableConsole:  true,
		EnableFile:     true,
		TimeFormat:     time.RFC3339,
		EnableCaller:   true,
		EnableStackTrace: false,
	}
}

// Initialize initializes the global logger with the given configuration
func Initialize(config Config) error {
	// Set time format
	zerolog.TimeFieldFormat = config.TimeFormat

	// Set log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure logger
	var writers []io.Writer

	// Console writer
	if config.EnableConsole {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: config.TimeFormat,
			NoColor:    false,
		}
		writers = append(writers, consoleWriter)
	}

	// File writer
	if config.EnableFile {
		// Create logs directory if it doesn't exist
		if err := os.MkdirAll("logs", 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(config.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writers = append(writers, file)
	}

	// Multi-writer
	multiWriter := zerolog.MultiLevelWriter(writers...)

	// Create logger
	Logger = zerolog.New(multiWriter).With().Timestamp().Logger()

	// Add caller information if enabled
	if config.EnableCaller {
		Logger = Logger.With().Caller().Logger()
	}

	// Add stack trace for error level if enabled
	if config.EnableStackTrace {
		zerolog.ErrorStackMarshaler = MarshalStack
	}

	// Set global logger
	log.Logger = Logger

	return nil
}

// MarshalStack marshals the stack trace for error logging
func MarshalStack(err error) interface{} {
	return err
}

// GetLogger returns the global logger instance
func GetLogger() zerolog.Logger {
	if zerolog.GlobalLevel() == zerolog.NoLevel {
		Initialize(DefaultConfig())
	}
	return Logger
}

// WithFields creates a logger with additional fields
func WithFields(fields map[string]interface{}) zerolog.Logger {
	logger := GetLogger()
	ctx := logger.With()
	
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}
	
	return ctx.Logger()
}

// WithRequest creates a logger with request context
func WithRequest(requestID, userID, organizationID, method, path string) zerolog.Logger {
	logger := GetLogger()
	return logger.With().
		Str("request_id", requestID).
		Str("user_id", userID).
		Str("organization_id", organizationID).
		Str("method", method).
		Str("path", path).
		Logger()
}

// WithError creates a logger with error context
func WithError(err error) zerolog.Logger {
	logger := GetLogger()
	return logger.With().Err(err).Logger()
}

// Info logs an info message
func Info(msg string, fields ...map[string]interface{}) {
	logger := GetLogger()
	if len(fields) > 0 {
		logger = WithFields(fields[0])
	}
	logger.Info().Msg(msg)
}

// Debug logs a debug message
func Debug(msg string, fields ...map[string]interface{}) {
	logger := GetLogger()
	if len(fields) > 0 {
		logger = WithFields(fields[0])
	}
	logger.Debug().Msg(msg)
}

// Warn logs a warning message
func Warn(msg string, fields ...map[string]interface{}) {
	logger := GetLogger()
	if len(fields) > 0 {
		logger = WithFields(fields[0])
	}
	logger.Warn().Msg(msg)
}

// Error logs an error message
func Error(msg string, err error, fields ...map[string]interface{}) {
	logger := GetLogger()
	logger = WithError(err)
	if len(fields) > 0 {
		logger = WithFields(fields[0])
	}
	logger.Error().Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, err error, fields ...map[string]interface{}) {
	logger := GetLogger()
	logger = WithError(err)
	if len(fields) > 0 {
		logger = WithFields(fields[0])
	}
	logger.Fatal().Msg(msg)
}

// LogDBQuery logs database query information
func LogDBQuery(query string, duration time.Duration, err error) {
	fields := map[string]interface{}{
		"query":    query,
		"duration": duration.Milliseconds(),
		"type":     "database",
	}
	
	if err != nil {
		Error("Database query failed", err, fields)
	} else {
		Debug("Database query executed", fields)
	}
}

// LogHTTPRequest logs HTTP request information
func LogHTTPRequest(method, path, userAgent, ipAddress string, statusCode int, duration time.Duration) {
	fields := map[string]interface{}{
		"method":      method,
		"path":        path,
		"user_agent":  userAgent,
		"ip_address":  ipAddress,
		"status_code": statusCode,
		"duration":    duration.Milliseconds(),
		"type":        "http_request",
	}
	
	if statusCode >= 400 {
		Warn("HTTP request completed with error", fields)
	} else {
		Info("HTTP request completed", fields)
	}
}

// LogAuthEvent logs authentication events
func LogAuthEvent(event, userID, organizationID, ipAddress string, success bool) {
	fields := map[string]interface{}{
		"event":          event,
		"user_id":        userID,
		"organization_id": organizationID,
		"ip_address":     ipAddress,
		"success":        success,
		"type":           "auth",
	}
	
	if success {
		Info("Authentication event", fields)
	} else {
		Warn("Authentication failed", fields)
	}
}

// LogBusinessEvent logs business events
func LogBusinessEvent(event, entityType, entityID, userID, organizationID string, data map[string]interface{}) {
	fields := map[string]interface{}{
		"event":          event,
		"entity_type":    entityType,
		"entity_id":      entityID,
		"user_id":        userID,
		"organization_id": organizationID,
		"type":           "business",
	}
	
	// Merge additional data
	for k, v := range data {
		fields[k] = v
	}
	
	Info("Business event", fields)
}

// LogSystemEvent logs system events
func LogSystemEvent(event, component string, data map[string]interface{}) {
	fields := map[string]interface{}{
		"event":     event,
		"component": component,
		"type":      "system",
	}
	
	// Merge additional data
	for k, v := range data {
		fields[k] = v
	}
	
	Info("System event", fields)
}