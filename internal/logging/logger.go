package logging

import (
	"github.com/sirupsen/logrus"
)

// Logger represents a structured logger that wraps logrus.Logger.
type Logger struct {
	*logrus.Logger
}

// NewLogger creates a new Logger instance with the specified log level.
func NewLogger(level string) *Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{}) // Log in JSON format for structured logs
	l.SetLevel(parseLogLevel(level))        // Set the logging level based on the string input
	return &Logger{l}
}

// SetLevel allows changing the logging level dynamically.
func (l *Logger) SetLevel(level string) {
	l.Logger.SetLevel(parseLogLevel(level))
}

// WithFields creates a new log entry with additional structured fields.
func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// WithError creates a new log entry that includes error information.
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// Fatalf logs a fatal error with a formatted message and then exits.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Logger.Fatalf(format, v...)
}

// Fatal logs a fatal error message and then exits.
func (l *Logger) Fatal(msg string) {
	l.Logger.Fatal(msg)
}

// parseLogLevel converts a string log level to the corresponding logrus log level.
func parseLogLevel(level string) logrus.Level {
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel // Default to info if the level is unknown
	}
}
