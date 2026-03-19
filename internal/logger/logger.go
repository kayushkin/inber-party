package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l Level) String() string {
	switch l {
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

type Logger struct {
	level  Level
	output *log.Logger
}

type LogEntry struct {
	Timestamp string            `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    string            `json:"caller,omitempty"`
}

var defaultLogger *Logger

func init() {
	defaultLogger = New(INFO)
}

func New(level Level) *Logger {
	return &Logger{
		level:  level,
		output: log.New(os.Stdout, "", 0),
	}
}

func SetLevel(level Level) {
	defaultLogger.level = level
}

func (l *Logger) log(level Level, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
		Fields:    fields,
	}

	// Add caller information for non-INFO level logs
	if level != INFO {
		if pc, file, line, ok := runtime.Caller(3); ok {
			fn := runtime.FuncForPC(pc)
			fnName := ""
			if fn != nil {
				fnName = fn.Name()
				// Get just the function name, not the full package path
				if idx := strings.LastIndex(fnName, "."); idx != -1 {
					fnName = fnName[idx+1:]
				}
			}
			// Get just the filename, not the full path
			if idx := strings.LastIndex(file, "/"); idx != -1 {
				file = file[idx+1:]
			}
			entry.Caller = fmt.Sprintf("%s:%d %s", file, line, fnName)
		}
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to basic logging if JSON marshaling fails
		l.output.Printf("LOG_ERROR: %v | %s: %s", err, level.String(), message)
		return
	}

	l.output.Println(string(jsonBytes))

	if level == FATAL {
		os.Exit(1)
	}
}

func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(DEBUG, message, f)
}

func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(INFO, message, f)
}

func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(WARN, message, f)
}

func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ERROR, message, f)
}

func (l *Logger) Fatal(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(FATAL, message, f)
}

// Convenience functions for default logger
func Debug(message string, fields ...map[string]interface{}) {
	defaultLogger.Debug(message, fields...)
}

func Info(message string, fields ...map[string]interface{}) {
	defaultLogger.Info(message, fields...)
}

func Warn(message string, fields ...map[string]interface{}) {
	defaultLogger.Warn(message, fields...)
}

func Error(message string, fields ...map[string]interface{}) {
	defaultLogger.Error(message, fields...)
}

func Fatal(message string, fields ...map[string]interface{}) {
	defaultLogger.Fatal(message, fields...)
}

// Printf-style convenience methods for easy migration from log.Printf
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debug(fmt.Sprintf(format, args...))
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Info(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warn(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Error(fmt.Sprintf(format, args...))
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatal(fmt.Sprintf(format, args...))
}

// Environment-based level setting
func SetLevelFromEnv() {
	envLevel := os.Getenv("LOG_LEVEL")
	switch strings.ToUpper(envLevel) {
	case "DEBUG":
		SetLevel(DEBUG)
	case "INFO", "":
		SetLevel(INFO)
	case "WARN":
		SetLevel(WARN)
	case "ERROR":
		SetLevel(ERROR)
	}
}