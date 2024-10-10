package logging

import (
	"fmt"
	"log"
	"os"
)

// Logger represents a structured logger.
type Logger struct {
	level string // Log level (e.g., debug, info, warn, error)
}

// NewLogger creates a new Logger instance with the specified log level.
func NewLogger(level string) *Logger {
	return &Logger{
		level: level,
	}
}

// SetLevel sets the logging level for the logger.
func (l *Logger) SetLevel(level string) {
	l.level = level
}

// Info logs an informational message.
func (l *Logger) Info(msg string) {
	if l.level == "debug" || l.level == "info" {
		log.Println("[INFO]", msg)
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string) {
	if l.level == "debug" {
		log.Println("[DEBUG]", msg)
	}
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string) {
	if l.level == "debug" || l.level == "info" || l.level == "warn" {
		log.Println("[WARN]", msg)
	}
}

// Error logs an error message.
func (l *Logger) Error(msg string) {
	log.Println("[ERROR]", msg)
}

// WithFields creates a new logger instance with additional fields.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	fieldString := ""
	for k, v := range fields {
		fieldString += fmt.Sprintf("%s=%v ", k, v)
	}
	return &Logger{
		level: l.level,
	}
}

// Fatalf logs a message at Fatal level and then exits the program
func (l *Logger) Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Fatal(msg) // or l.LogFatal(msg) if you have a method for logging fatal errors
}

// WithError creates a new logger instance including error information.
func (l *Logger) WithError(err error) *Logger {
	return l.WithFields(map[string]interface{}{"error": err.Error()})
}

// Fatal logs a fatal message and exits the application.
func (l *Logger) Fatal(msg string) {
	log.Fatalln("[FATAL]", msg)
}

func init() {
	// Set log output to standard output
	log.SetOutput(os.Stdout)
}
