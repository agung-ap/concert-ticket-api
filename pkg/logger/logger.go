package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents a logging level
type LogLevel int

const (
	// DEBUG represents debug level
	DEBUG LogLevel = iota
	// INFO represents info level
	INFO
	// WARN represents warn level
	WARN
	// ERROR represents error level
	ERROR
	// FATAL represents fatal level
	FATAL
)

// Logger represents a logger
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
}

type logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger
func NewLogger(level string) Logger {
	return &logger{
		level:  parseLevel(level),
		logger: log.New(os.Stdout, "", 0),
	}
}

// parseLevel parses a string level to a LogLevel
func parseLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

// levelToString converts a LogLevel to a string
func levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// log logs a message with the given level
func (l *logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	levelStr := levelToString(level)
	message := fmt.Sprintf(format, args...)

	l.logger.Printf("%s [%s] %s", now, levelStr, message)
}

// Debug logs a debug message
func (l *logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits
func (l *logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}
