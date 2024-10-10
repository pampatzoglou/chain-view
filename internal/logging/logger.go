package logging

import (
	"github.com/sirupsen/logrus"
)

// Logger represents a structured logger.
type Logger struct {
	*logrus.Logger
}

// NewLogger creates a new Logger instance with the specified log level.
func NewLogger(level string) *Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	l.SetLevel(parseLogLevel(level))
	return &Logger{l}
}

// SetLevel sets the logging level for the logger.
func (l *Logger) SetLevel(level string) {
	l.Logger.SetLevel(parseLogLevel(level))
}

// WithFields creates a new entry with additional fields.
func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// WithError creates a new entry with error information.
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// parseLogLevel converts a string log level to logrus.Level
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
		return logrus.InfoLevel
	}
}
