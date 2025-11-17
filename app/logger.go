package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

	// File logging
	logFile            *os.File
	logFileMutex       sync.Mutex
	logWriter          io.Writer
	currentDate        string
	fileLoggingEnabled bool
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

// isBuildMode checks if the app is running in build/production mode
// In dev mode (wails dev), the executable is typically in a temp directory or project root
// In build mode, the executable is in the build directory or installed location
func isBuildMode() bool {
	execPath, err := os.Executable()
	if err != nil {
		return false
	}

	// Resolve symlinks to get real path
	realPath, err := filepath.EvalSymlinks(execPath)
	if err == nil {
		execPath = realPath
	}

	// Get absolute path
	absPath, err := filepath.Abs(execPath)
	if err != nil {
		return false
	}

	// Check if running from build directory
	// Build mode: executable is in build/bin/ or similar
	// Dev mode: executable is typically in a temp directory or project root
	absPath = strings.ToLower(absPath)

	// Check for common build indicators
	buildIndicators := []string{
		"build/bin",
		"build/",
		".app/contents/macos", // macOS app bundle
		"application support",
		"program files",  // Windows
		"/usr/local/bin", // Linux installed
		"/opt/",          // Linux installed
	}

	// Check if path contains any build indicators
	for _, indicator := range buildIndicators {
		if strings.Contains(absPath, indicator) {
			return true
		}
	}

	// Check if NOT in common dev locations
	devIndicators := []string{
		"/tmp/",
		"/var/folders/", // macOS temp
		"go-build",      // Go temp build
		"wails",         // Wails temp
	}

	for _, indicator := range devIndicators {
		if strings.Contains(absPath, indicator) {
			return false
		}
	}

	// Additional check: if executable name matches and is not in temp, likely build mode
	// In dev mode, Wails often uses temp executables
	if runtime.GOOS == "darwin" {
		// On macOS, if it's an .app bundle, it's definitely build mode
		if strings.HasSuffix(absPath, ".app/contents/macos/cloudflared-desktop-tunnel") {
			return true
		}
	}

	// Default: assume dev mode if uncertain
	return false
}

// getLogDir returns the path to the log directory
func getLogDir() (string, error) {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// Create app log directory if it doesn't exist
	logDir := filepath.Join(configDir, "cloudflared-desktop-tunnel", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", err
	}

	return logDir, nil
}

// getLogFilePath returns the path to today's log file
func getLogFilePath() (string, error) {
	logDir, err := getLogDir()
	if err != nil {
		return "", err
	}

	dateStr := time.Now().Format("2006-01-02")
	return filepath.Join(logDir, fmt.Sprintf("app-%s.log", dateStr)), nil
}

// ensureLogFile ensures the log file is open and rotated if needed
func ensureLogFile() error {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	today := time.Now().Format("2006-01-02")

	// If file is already open for today, no need to change
	if logFile != nil && currentDate == today {
		return nil
	}

	// Close existing file if open
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}

	// Get today's log file path
	logPath, err := getLogFilePath()
	if err != nil {
		return err
	}

	// Open log file in append mode
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	logFile = file
	currentDate = today

	// Create multi-writer: both stderr and file
	logWriter = io.MultiWriter(os.Stderr, logFile)

	// Set log output to our multi-writer
	log.SetOutput(logWriter)

	return nil
}

// InitFileLogging initializes file logging only in build mode
func InitFileLogging() error {
	// Only enable file logging in build mode
	if !isBuildMode() {
		fileLoggingEnabled = false
		return nil
	}

	fileLoggingEnabled = true
	if err := ensureLogFile(); err != nil {
		return fmt.Errorf("failed to initialize file logging: %w", err)
	}
	return nil
}

// CloseFileLogging closes the log file and deletes it
func CloseFileLogging() {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	if logFile != nil {
		// Get the log file path before closing
		logPath := logFile.Name()

		// Close the file
		logFile.Close()
		logFile = nil
		logWriter = nil

		// Reset log output to stderr
		log.SetOutput(os.Stderr)

		// Delete the log file
		if logPath != "" {
			os.Remove(logPath)
		}
	} else if currentDate != "" {
		// If file was already closed but we have the date, try to delete today's log file
		if logPath, err := getLogFilePath(); err == nil {
			os.Remove(logPath)
		}
	}
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

// writeLog writes a log message, ensuring file is open
func (l *Logger) writeLog(msg string) {
	// If file logging is not enabled (dev mode), just write to stderr
	if !fileLoggingEnabled {
		log.SetOutput(os.Stderr)
		log.Println(msg)
		return
	}

	// Ensure log file is open (will check date and rotate if needed)
	if err := ensureLogFile(); err != nil {
		// If file logging fails, fall back to stderr only
		log.SetOutput(os.Stderr)
		log.Println(msg)
		return
	}

	log.Println(msg)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if msg := l.formatMessage(LevelDebug, format, args...); msg != "" {
		l.writeLog(msg)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if msg := l.formatMessage(LevelInfo, format, args...); msg != "" {
		l.writeLog(msg)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if msg := l.formatMessage(LevelWarn, format, args...); msg != "" {
		l.writeLog(msg)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if msg := l.formatMessage(LevelError, format, args...); msg != "" {
		l.writeLog(msg)
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
