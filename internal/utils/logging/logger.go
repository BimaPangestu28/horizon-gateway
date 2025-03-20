package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel string

const (
	// LevelDebug is for detailed troubleshooting information
	LevelDebug LogLevel = "debug"

	// LevelInfo is for general operational information
	LevelInfo LogLevel = "info"

	// LevelWarn is for potentially harmful situations
	LevelWarn LogLevel = "warn"

	// LevelError is for error events that might still allow the application to continue
	LevelError LogLevel = "error"

	// LevelFatal is for very severe error events that will presumably lead the application to abort
	LevelFatal LogLevel = "fatal"
)

// Logger interface defines methods for logging at different levels
type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	Fatal(msg string, keyvals ...interface{})
}

// JSONLogger implements the Logger interface using JSON structured logging
type JSONLogger struct {
	level  LogLevel
	output *os.File
}

// LogEntry represents a single log entry in JSON format
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Fields    interface{} `json:"fields,omitempty"`
}

// NewLogger creates a new JSON logger with default settings
func NewLogger() Logger {
	return &JSONLogger{
		level:  LevelInfo,
		output: os.Stdout,
	}
}

// Debug logs a message at debug level
func (l *JSONLogger) Debug(msg string, keyvals ...interface{}) {
	if l.level == LevelDebug {
		l.log(LevelDebug, msg, keyvals...)
	}
}

// Info logs a message at info level
func (l *JSONLogger) Info(msg string, keyvals ...interface{}) {
	if l.level != LevelError && l.level != LevelFatal {
		l.log(LevelInfo, msg, keyvals...)
	}
}

// Warn logs a message at warn level
func (l *JSONLogger) Warn(msg string, keyvals ...interface{}) {
	if l.level != LevelError && l.level != LevelFatal {
		l.log(LevelWarn, msg, keyvals...)
	}
}

// Error logs a message at error level
func (l *JSONLogger) Error(msg string, keyvals ...interface{}) {
	l.log(LevelError, msg, keyvals...)
}

// Fatal logs a message at fatal level and then exits the program
func (l *JSONLogger) Fatal(msg string, keyvals ...interface{}) {
	l.log(LevelFatal, msg, keyvals...)
	os.Exit(1)
}

// log writes a log entry to the output
func (l *JSONLogger) log(level LogLevel, msg string, keyvals ...interface{}) {
	// Create fields from key-value pairs
	fields := make(map[string]interface{})

	// Process key-value pairs
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key, ok := keyvals[i].(string)
			if !ok {
				key = fmt.Sprintf("%v", keyvals[i])
			}
			fields[key] = keyvals[i+1]
		} else {
			// Handle odd number of arguments
			fields["_dangling"] = keyvals[i]
		}
	}

	// Create log entry
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     string(level),
		Message:   msg,
		Fields:    fields,
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling log entry: %v\n", err)
		return
	}

	// Write to output
	fmt.Fprintln(l.output, string(data))
}
