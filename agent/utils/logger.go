package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger provides structured logging for the agent
type Logger struct {
	level   LogLevel
	verbose bool
	logger  *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(verbose bool) *Logger {
	level := INFO
	if verbose {
		level = DEBUG
	}

	return &Logger{
		level:   level,
		verbose: verbose,
		logger:  log.New(os.Stdout, "", 0),
	}
}

// Debug logs debug messages
func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	if l.level <= DEBUG {
		l.log("DEBUG", msg, keyvals...)
	}
}

// Info logs info messages
func (l *Logger) Info(msg string, keyvals ...interface{}) {
	if l.level <= INFO {
		l.log("INFO", msg, keyvals...)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	if l.level <= WARN {
		l.log("WARN", msg, keyvals...)
	}
}

// Error logs error messages
func (l *Logger) Error(msg string, keyvals ...interface{}) {
	if l.level <= ERROR {
		l.log("ERROR", msg, keyvals...)
	}
}

// log formats and outputs log messages
func (l *Logger) log(level, msg string, keyvals ...interface{}) {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")
	
	var parts []string
	parts = append(parts, fmt.Sprintf("time=%s", timestamp))
	parts = append(parts, fmt.Sprintf("level=%s", level))
	parts = append(parts, fmt.Sprintf("msg=\"%s\"", msg))
	
	// Add key-value pairs
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := fmt.Sprintf("%v", keyvals[i])
			value := fmt.Sprintf("%v", keyvals[i+1])
			parts = append(parts, fmt.Sprintf("%s=\"%s\"", key, value))
		}
	}
	
	l.logger.Println(strings.Join(parts, " "))
}