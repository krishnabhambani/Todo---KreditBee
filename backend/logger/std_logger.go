package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// stdLogger is the stdlib-based implementation of Logger.
// It produces lines in the format:
//
//	[LEVEL] 2006-01-02T15:04:05Z07:00 | message | key=value key=value ...
type stdLogger struct {
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
	fatalLog *log.Logger
}

// NewLogger returns a Logger backed by Go's stdlib log package.
// All levels write to stderr to keep stdout clean for health checks / JSON output.
func NewLogger() Logger {
	flags := 0 // We handle timestamps ourselves for consistent formatting.
	return &stdLogger{
		infoLog:  log.New(os.Stderr, "", flags),
		warnLog:  log.New(os.Stderr, "", flags),
		errorLog: log.New(os.Stderr, "", flags),
		debugLog: log.New(os.Stderr, "", flags),
		fatalLog: log.New(os.Stderr, "", flags),
	}
}

// format builds the final log line string.
func format(level, msg string, fields []Field) string {
	ts := time.Now().UTC().Format(time.RFC3339)
	b := &strings.Builder{}
	fmt.Fprintf(b, "[%s] %s | %s", level, ts, msg)
	for _, f := range fields {
		fmt.Fprintf(b, " | %s=%v", f.Key, f.Value)
	}
	return b.String()
}

func (l *stdLogger) Info(msg string, fields ...Field) {
	l.infoLog.Println(format("INFO ", msg, fields))
}

func (l *stdLogger) Warn(msg string, fields ...Field) {
	l.warnLog.Println(format("WARN ", msg, fields))
}

func (l *stdLogger) Error(msg string, fields ...Field) {
	l.errorLog.Println(format("ERROR", msg, fields))
}

func (l *stdLogger) Debug(msg string, fields ...Field) {
	l.debugLog.Println(format("DEBUG", msg, fields))
}

func (l *stdLogger) Fatal(msg string, fields ...Field) {
	l.fatalLog.Println(format("FATAL", msg, fields))
	os.Exit(1)
}
