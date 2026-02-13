package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog for structured logging
type Logger struct {
	*zerolog.Logger
}

// New creates a new Logger instance
func New(level string) *Logger {
	// Parse log level
	var zerologLevel zerolog.Level
	switch level {
	case "debug":
		zerologLevel = zerolog.DebugLevel
	case "info":
		zerologLevel = zerolog.InfoLevel
	case "warn":
		zerologLevel = zerolog.WarnLevel
	case "error":
		zerologLevel = zerolog.ErrorLevel
	default:
		zerologLevel = zerolog.InfoLevel
	}

	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(zerologLevel)

	// Add console output for development
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}

	logger := zerolog.New(consoleWriter).
		With().
		Timestamp().
		Caller().
		Str("service", "pipeline-arch").
		Logger()

	return &Logger{Logger: &logger}
}

// WithComponent returns a new logger with a component field
func (l *Logger) WithComponent(component string) *zerolog.Logger {
	return l.Logger.With().Str("component", component).Logger()
}

// WithRequestID returns a new logger with a request ID field
func (l *Logger) WithRequestID(requestID string) *zerolog.Logger {
	return l.Logger.With().Str("request_id", requestID).Logger()
}

// WithUser returns a new logger with a user field
func (l *Logger) WithUser(userID string) *zerolog.Logger {
	return l.Logger.With().Str("user_id", userID).Logger()
}

// Debug logs a debug message
func (l *Logger) Debug() *zerolog.Event {
	return l.Logger.Debug()
}

// Info logs an info message
func (l *Logger) Info() *zerolog.Event {
	return l.Logger.Info()
}

// Warn logs a warning message
func (l *Logger) Warn() *zerolog.Event {
	return l.Logger.Warn()
}

// Error logs an error message
func (l *Logger) Error() *zerolog.Event {
	return l.Logger.Error()
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal() *zerolog.Event {
	return l.Logger.Fatal()
}

// Panic logs a panic message and panics
func (l *Logger) Panic() *zerolog.Event {
	return l.Logger.Panic()
}

// WithErr logs an error
func (l *Logger) WithErr(err error) *zerolog.Event {
	return l.Logger.With().Err(err).Logger()
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *zerolog.Logger {
	return l.Logger.With().Interface(key, value).Logger()
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *zerolog.Logger {
	logger := l.Logger
	for key, value := range fields {
		logger = logger.With().Interface(key, value).Logger()
	}
	return &logger
}

// Named returns a new logger with the given name
func (l *Logger) Named(name string) *Logger {
	return &Logger{Logger: l.Logger.Named(name)}
}