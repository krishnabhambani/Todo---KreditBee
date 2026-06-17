package logger

// Field is a structured key-value pair attached to a log entry.
type Field struct {
	Key   string
	Value any
}

// F is a convenience constructor for a Field.
func F(key string, val any) Field {
	return Field{Key: key, Value: val}
}

// Logger is the application-wide structured logging contract.
// All consumers depend on this interface, never on a concrete type.
type Logger interface {
	// Info logs a message at INFO level — normal operational events.
	Info(msg string, fields ...Field)

	// Warn logs a message at WARN level — unexpected but non-fatal conditions.
	Warn(msg string, fields ...Field)

	// Error logs a message at ERROR level — failures that need attention.
	Error(msg string, fields ...Field)

	// Debug logs a message at DEBUG level — verbose development details.
	Debug(msg string, fields ...Field)

	// Fatal logs at FATAL level then calls os.Exit(1).
	Fatal(msg string, fields ...Field)
}
