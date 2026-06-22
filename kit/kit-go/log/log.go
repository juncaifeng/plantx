// Package log provides a structured logging abstraction for Kit services.
package log

import "context"

// Logger is a structured logging abstraction.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
}

// Field is a key-value log field.
type Field struct {
	Key   string
	Value any
}

// F builds a Field.
func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// FromContext returns a logger enriched with request context.
func FromContext(ctx context.Context) Logger {
	if l, ok := ctx.Value(loggerKey).(Logger); ok {
		return l
	}
	return nopLogger{}
}

// WithContext injects a logger into the context.
func WithContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

type loggerKeyType struct{}

var loggerKey loggerKeyType

type nopLogger struct{}

func (nopLogger) Debug(string, ...Field) {}
func (nopLogger) Info(string, ...Field)  {}
func (nopLogger) Warn(string, ...Field)  {}
func (nopLogger) Error(string, ...Field) {}
func (n nopLogger) With(...Field) Logger { return n }
