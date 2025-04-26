package logger

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	apihelpers "github.com/JalajGoswami/video-ad-metrics/internal/api-helpers"
)

type LogLevel uint
type LogColor string

// Colors for different log levels
const (
	ColorReset  LogColor = "\033[0m"
	ColorRed    LogColor = "\033[31m"
	ColorGreen  LogColor = "\033[32m"
	ColorYellow LogColor = "\033[33m"
	ColorBlue   LogColor = "\033[34m"
)

var logLevelNames = []string{"DEBUG", "INFO", "ERROR"}
var logLevelColors = []LogColor{ColorYellow, ColorGreen, ColorRed}

// Log levels
const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelError
)

func (l LogLevel) String() string {
	return logLevelNames[l]
}

// getColorForLevel returns the ANSI color code for the given log level
func (l LogLevel) Color() LogColor {
	return logLevelColors[l]
}

func getLogLevelFromEnv() LogLevel {
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "DEBUG", "debug":
		return LevelDebug
	case "INFO", "info":
		return LevelInfo
	case "ERROR", "error":
		return LevelError
	}
	return LevelDebug
}

// requestLogger is a structured logger for HTTP requests
type requestLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	logLevel    LogLevel
}

var RequestLogger *requestLogger

// New creates a new Request Logger instance
func SetupRequestLogger() {
	RequestLogger = &requestLogger{
		infoLogger:  log.New(os.Stdout, "", 0),
		errorLogger: log.New(os.Stderr, "", 0),
		debugLogger: log.New(os.Stdout, "", 0),
		logLevel:    getLogLevelFromEnv(),
	}
}

// formatStructuredLog formats a structured log message with the provided information
func formatStructuredLog(level LogLevel, traceID, method, url, message string) string {
	color := level.Color()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%s%s:%s %s %s %s %s %s",
		color, level, ColorReset,
		traceID, timestamp, method, url, message)
}

// Info logs an informational message for an HTTP request
func (l *requestLogger) Info(r *http.Request, message string, args ...any) {
	if l.logLevel > LevelInfo {
		return
	}
	logMessage := formatStructuredLog(LevelInfo, apihelpers.GetTraceId(r), r.Method, r.URL.String(), fmt.Sprintf(message, args...))
	l.infoLogger.Println(logMessage)
}

// Error logs an error message for an HTTP request
func (l *requestLogger) Error(r *http.Request, message string, args ...any) {
	if l.logLevel > LevelError {
		return
	}
	logMessage := formatStructuredLog(LevelError, apihelpers.GetTraceId(r), r.Method, r.URL.String(), fmt.Sprintf(message, args...))
	l.errorLogger.Println(logMessage)
}

// Debug logs a debug message for an HTTP request
func (l *requestLogger) Debug(r *http.Request, message string, args ...any) {
	if l.logLevel > LevelDebug {
		return
	}
	logMessage := formatStructuredLog(LevelDebug, apihelpers.GetTraceId(r), r.Method, r.URL.String(), fmt.Sprintf(message, args...))
	l.debugLogger.Println(logMessage)
}

// LoggingMiddleware wraps HandleFunc with logging
func (l *requestLogger) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.Info(r, "")
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Non-structured logging functions for scripts and non-request scenarios

// InfoLog logs a simple informational message without structure
func InfoLog(message string, args ...any) {
	fmt.Printf("%s%s:%s %s\n", ColorGreen, LevelInfo, ColorReset, fmt.Sprintf(message, args...))
}

// ErrorLog logs a simple error message without structure
func ErrorLog(message string, args ...any) {
	fmt.Printf("%s%s:%s %s\n", ColorRed, LevelError, ColorReset, fmt.Sprintf(message, args...))
}

// FatalLog logs a simple fatal message without structure that will exit the program
func FatalLog(message string, args ...any) {
	fmt.Printf("%s%s:%s %s\n", ColorRed, LevelError, ColorReset, fmt.Sprintf(message, args...))
	os.Exit(1)
}

// DebugLog logs a simple debug message without structure
func DebugLog(message string, args ...any) {
	fmt.Printf("%s%s:%s %s\n", ColorYellow, LevelDebug, ColorReset, fmt.Sprintf(message, args...))
}

func LogColored(color LogColor, message string, args ...any) {
	fmt.Printf("%s%s%s\n", color, fmt.Sprintf(message, args...), ColorReset)
}
