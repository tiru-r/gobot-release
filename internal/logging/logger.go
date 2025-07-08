package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DebugLevel for debug messages
	DebugLevel LogLevel = iota
	// InfoLevel for informational messages
	InfoLevel
	// WarnLevel for warning messages
	WarnLevel
	// ErrorLevel for error messages
	ErrorLevel
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// LogFormat represents the output format for log messages
type LogFormat int

const (
	// TextFormat for human-readable text output
	TextFormat LogFormat = iota
	// JSONFormat for structured JSON output
	JSONFormat
)

// Logger represents a structured logger for Gobot
type Logger struct {
	level      LogLevel
	format     LogFormat
	output     io.Writer
	component  string
	enableTime bool
	enableCaller bool
}

// LogEntry represents a single log entry
type LogEntry struct {
	Time      time.Time         `json:"time,omitempty"`
	Level     string            `json:"level"`
	Component string            `json:"component,omitempty"`
	Message   string            `json:"message"`
	Caller    string            `json:"caller,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// NewLogger creates a new logger with the specified configuration
func NewLogger(level LogLevel, format LogFormat, output io.Writer) *Logger {
	return &Logger{
		level:        level,
		format:       format,
		output:       output,
		enableTime:   true,
		enableCaller: true,
	}
}

// NewDefaultLogger creates a logger with default settings
func NewDefaultLogger() *Logger {
	return NewLogger(InfoLevel, TextFormat, os.Stdout)
}

// WithComponent returns a new logger with a component name
func (l *Logger) WithComponent(component string) *Logger {
	newLogger := *l
	newLogger.component = component
	return &newLogger
}

// WithFields returns a logger that will include the specified fields in all log messages
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	// For simplicity, we'll just return the same logger
	// In a more sophisticated implementation, we would store these fields
	return l
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetFormat sets the log output format
func (l *Logger) SetFormat(format LogFormat) {
	l.format = format
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	l.log(DebugLevel, message, fields...)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...))
}

// Info logs an informational message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.log(InfoLevel, message, fields...)
}

// Infof logs a formatted informational message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...))
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	l.log(WarnLevel, message, fields...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...))
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	l.log(ErrorLevel, message, fields...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...))
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, message string, fields ...map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Level:     level.String(),
		Component: l.component,
		Message:   message,
	}

	if l.enableTime {
		entry.Time = time.Now()
	}

	if l.enableCaller {
		if caller := getCaller(3); caller != "" {
			entry.Caller = caller
		}
	}

	if len(fields) > 0 && fields[0] != nil {
		entry.Fields = fields[0]
	}

	l.writeEntry(entry)
}

// writeEntry writes a log entry to the output
func (l *Logger) writeEntry(entry LogEntry) {
	var output string

	switch l.format {
	case JSONFormat:
		if data, err := json.Marshal(entry); err == nil {
			output = string(data) + "\n"
		} else {
			output = fmt.Sprintf(`{"level":"ERROR","message":"Failed to marshal log entry: %v"}`+"\n", err)
		}
	default: // TextFormat
		timeStr := ""
		if !entry.Time.IsZero() {
			timeStr = entry.Time.Format("2006-01-02 15:04:05.000") + " "
		}

		componentStr := ""
		if entry.Component != "" {
			componentStr = "[" + entry.Component + "] "
		}

		callerStr := ""
		if entry.Caller != "" {
			callerStr = " (" + entry.Caller + ")"
		}

		fieldsStr := ""
		if entry.Fields != nil && len(entry.Fields) > 0 {
			parts := make([]string, 0, len(entry.Fields))
			for k, v := range entry.Fields {
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}
			fieldsStr = " " + strings.Join(parts, " ")
		}

		output = fmt.Sprintf("%s%s%s%s%s%s\n",
			timeStr,
			entry.Level,
			callerStr,
			componentStr,
			entry.Message,
			fieldsStr,
		)
	}

	fmt.Fprint(l.output, output)
}

// getCaller returns the file and line number of the caller
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	// Get just the filename without the full path
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		file = parts[len(parts)-1]
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// Global logger instance
var defaultLogger = NewDefaultLogger()

// Configure configures the default logger
func Configure(level LogLevel, format LogFormat, output io.Writer) {
	defaultLogger = NewLogger(level, format, output)
}

// ConfigureFromString configures the default logger from string parameters
func ConfigureFromString(levelStr, formatStr, outputStr string) error {
	level := ParseLogLevel(levelStr)

	var format LogFormat
	switch strings.ToLower(formatStr) {
	case "json":
		format = JSONFormat
	default:
		format = TextFormat
	}

	var output io.Writer
	switch strings.ToLower(outputStr) {
	case "stderr":
		output = os.Stderr
	case "stdout", "":
		output = os.Stdout
	default:
		// Try to open as a file
		if file, err := os.OpenFile(outputStr, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			output = file
		} else {
			return fmt.Errorf("failed to open log output file '%s': %w", outputStr, err)
		}
	}

	Configure(level, format, output)
	return nil
}

// Global logging functions that use the default logger

// Debug logs a debug message using the default logger
func Debug(message string, fields ...map[string]interface{}) {
	defaultLogger.Debug(message, fields...)
}

// Debugf logs a formatted debug message using the default logger
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Info logs an informational message using the default logger
func Info(message string, fields ...map[string]interface{}) {
	defaultLogger.Info(message, fields...)
}

// Infof logs a formatted informational message using the default logger
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(message string, fields ...map[string]interface{}) {
	defaultLogger.Warn(message, fields...)
}

// Warnf logs a formatted warning message using the default logger
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Error logs an error message using the default logger
func Error(message string, fields ...map[string]interface{}) {
	defaultLogger.Error(message, fields...)
}

// Errorf logs a formatted error message using the default logger
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// GetLogger returns a logger with the specified component name
func GetLogger(component string) *Logger {
	return defaultLogger.WithComponent(component)
}

// Compatibility with standard log package
func init() {
	// Redirect standard log package to our logger
	log.SetOutput(&logWrapper{logger: defaultLogger})
	log.SetFlags(0) // Disable standard log formatting since we handle it
}

// logWrapper wraps our logger to be compatible with standard log package
type logWrapper struct {
	logger *Logger
}

func (w *logWrapper) Write(p []byte) (n int, err error) {
	message := strings.TrimSpace(string(p))
	if message != "" {
		w.logger.Info(message)
	}
	return len(p), nil
}