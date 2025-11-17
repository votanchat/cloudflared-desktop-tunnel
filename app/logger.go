package app

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	currentLevel LogLevel = LevelInfo
	logColors             = map[LogLevel]string{
		LevelDebug: "\033[36m", // Cyan
		LevelInfo:  "\033[32m", // Green
		LevelWarn:  "\033[33m", // Yellow
		LevelError: "\033[31m", // Red
	}
	resetColor = "\033[0m"
	levelNames = map[LogLevel]string{
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
	}
)

// Logger provides structured logging
type Logger struct {
	prefix string
}

// NewLogger creates a new logger with optional prefix
func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}

// SetLevel sets the global log level
func SetLevel(level LogLevel) {
	currentLevel = level
}

// formatMessage formats a log message with timestamp, level, and prefix
func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	if level < currentLevel {
		return ""
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]
	color := logColors[level]

	var prefixStr string
	if l.prefix != "" {
		prefixStr = fmt.Sprintf("[%s] ", l.prefix)
	}

	message := fmt.Sprintf(format, args...)

	// Only use colors if output is a terminal
	useColor := isTerminal()

	if useColor {
		return fmt.Sprintf("%s%s[%s]%s %s%s%s",
			timestamp, color, levelName, resetColor, prefixStr, message, resetColor)
	}

	return fmt.Sprintf("%s [%s] %s%s",
		timestamp, levelName, prefixStr, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if msg := l.formatMessage(LevelDebug, format, args...); msg != "" {
		log.Println(msg)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if msg := l.formatMessage(LevelInfo, format, args...); msg != "" {
		log.Println(msg)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if msg := l.formatMessage(LevelWarn, format, args...); msg != "" {
		log.Println(msg)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if msg := l.formatMessage(LevelError, format, args...); msg != "" {
		log.Println(msg)
	}
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// Package-level convenience functions
var (
	appLogger     = NewLogger("APP")
	tunnelLogger  = NewLogger("TUNNEL")
	backendLogger = NewLogger("BACKEND")
	serverLogger  = NewLogger("SERVER")
	binaryLogger  = NewLogger("BINARY")
)

// Convenience functions for package-level logging
func logDebug(format string, args ...interface{}) {
	appLogger.Debug(format, args...)
}

func logInfo(format string, args ...interface{}) {
	appLogger.Info(format, args...)
}

func logWarn(format string, args ...interface{}) {
	appLogger.Warn(format, args...)
}

func logError(format string, args ...interface{}) {
	appLogger.Error(format, args...)
}

// GetLogger returns a logger instance for a specific component
func GetLogger(component string) *Logger {
	return NewLogger(strings.ToUpper(component))
}
