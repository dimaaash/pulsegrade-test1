// Package logger provides a configurable logging facility that can be enabled or disabled
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// LogLevel represents the verbosity level of logging
type LogLevel int

// Available log levels
const (
	LevelNone  LogLevel = iota // No logging
	LevelError                 // Only errors
	LevelWarn                  // Warnings and errors
	LevelInfo                  // Informational messages, warnings, and errors
	LevelDebug                 // Debug messages, informational messages, warnings, and errors
)

var (
	// Default logger instance
	defaultLogger *Logger
	// once          sync.Once
)

// Logger wraps the standard log package with additional features
type Logger struct {
	enabled bool
	level   LogLevel
	logger  *log.Logger
	mu      sync.Mutex
}

// Config holds configuration for the logger
type Config struct {
	Enabled bool
	Level   LogLevel
	Output  io.Writer
}

// init initializes the default logger
func init() {
	defaultLogger = &Logger{
		enabled: true,
		level:   LevelInfo,
		logger:  log.New(os.Stdout, "", log.LstdFlags),
	}
}

// New creates a new logger with the provided configuration
func New(config Config) *Logger {
	output := config.Output
	if output == nil {
		output = os.Stdout
	}

	return &Logger{
		enabled: config.Enabled,
		level:   config.Level,
		logger:  log.New(output, "", log.LstdFlags),
	}
}

// SetDefault sets the default logger instance
func SetDefault(l *Logger) {
	defaultLogger = l
}

// Configure configures the default logger
func Configure(config Config) {
	output := config.Output
	if output == nil {
		output = os.Stdout
	}

	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	defaultLogger.enabled = config.Enabled
	defaultLogger.level = config.Level
	defaultLogger.logger = log.New(output, "", log.LstdFlags)
}

// // SetEnabled enables or disables logging for the default logger
// func SetEnabled(enabled bool) {
// 	defaultLogger.mu.Lock()
// 	defer defaultLogger.mu.Unlock()
// 	defaultLogger.enabled = enabled
// }

// // SetLevel sets the log level for the default logger
// func SetLevel(level LogLevel) {
// 	defaultLogger.mu.Lock()
// 	defer defaultLogger.mu.Unlock()
// 	defaultLogger.level = level
// }

// Debug logs a debug message if the logger is enabled and level is appropriate
func Debug(format string, v ...interface{}) {
	if defaultLogger.enabled && defaultLogger.level >= LevelDebug {
		defaultLogger.logger.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs an info message if the logger is enabled and level is appropriate
func Info(format string, v ...interface{}) {
	if defaultLogger.enabled && defaultLogger.level >= LevelInfo {
		defaultLogger.logger.Printf("[INFO] "+format, v...)
	}
}

// Warn logs a warning message if the logger is enabled and level is appropriate
func Warn(format string, v ...interface{}) {
	if defaultLogger.enabled && defaultLogger.level >= LevelWarn {
		defaultLogger.logger.Printf("[WARN] "+format, v...)
	}
}

// Error logs an error message if the logger is enabled and level is appropriate
func Error(format string, v ...interface{}) {
	if defaultLogger.enabled && defaultLogger.level >= LevelError {
		defaultLogger.logger.Printf("[ERROR] "+format, v...)
	}
}

// Fatal logs a fatal error message and exits
func Fatal(format string, v ...interface{}) {
	if defaultLogger.enabled {
		defaultLogger.logger.Fatalf("[FATAL] "+format, v...)
	}
	// Even if logging is disabled, we still need to exit
	os.Exit(1)
}

// Methods for Logger instance

// Debug logs a debug message using the logger instance
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.enabled && l.level >= LevelDebug {
		l.logger.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs an info message using the logger instance
func (l *Logger) Info(format string, v ...interface{}) {
	if l.enabled && l.level >= LevelInfo {
		l.logger.Printf("[INFO] "+format, v...)
	}
}

// Warn logs a warning message using the logger instance
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.enabled && l.level >= LevelWarn {
		l.logger.Printf("[WARN] "+format, v...)
	}
}

// Error logs an error message using the logger instance
func (l *Logger) Error(format string, v ...interface{}) {
	if l.enabled && l.level >= LevelError {
		l.logger.Printf("[ERROR] "+format, v...)
	}
}

// Fatal logs a fatal error message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.enabled {
		l.logger.Fatalf("[FATAL] "+format, v...)
	}
	// Even if logging is disabled, we still need to exit
	os.Exit(1)
}

// String returns a string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LevelNone:
		return "NONE"
	case LevelError:
		return "ERROR"
	case LevelWarn:
		return "WARN"
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	default:
		return fmt.Sprintf("LogLevel(%d)", l)
	}
}

// LevelFromString converts a string to a LogLevel
func LevelFromString(level string) LogLevel {
	switch level {
	case "NONE":
		return LevelNone
	case "ERROR":
		return LevelError
	case "WARN":
		return LevelWarn
	case "INFO":
		return LevelInfo
	case "DEBUG":
		return LevelDebug
	default:
		return LevelInfo // Default to INFO if not recognized
	}
}
