package zaplog

import (
	"context"

	kitctx "github.com/plantx/kit/kit-go/context"
	"github.com/plantx/kit/kit-go/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger to implement kit/log.Logger.
type Logger struct {
	zap *zap.Logger
}

// New creates a new zap-based logger.
func New(opts ...Option) (*Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	for _, o := range opts {
		cfg = o(cfg)
	}
	z, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return &Logger{zap: z}, nil
}

// Option customizes the zap config.
type Option func(zap.Config) zap.Config

// WithDevelopment switches to development config.
func WithDevelopment() Option {
	return func(c zap.Config) zap.Config {
		return zap.NewDevelopmentConfig()
	}
}

// WithLevel sets the minimum log level.
func WithLevel(lvl zapcore.Level) Option {
	return func(c zap.Config) zap.Config {
		c.Level = zap.NewAtomicLevelAt(lvl)
		return c
	}
}

func (l *Logger) fields(fs ...log.Field) []zap.Field {
	out := make([]zap.Field, len(fs))
	for i, f := range fs {
		out[i] = zap.Any(f.Key, f.Value)
	}
	return out
}

// Debug implements log.Logger.
func (l *Logger) Debug(msg string, fs ...log.Field) {
	l.zap.Debug(msg, l.fields(fs...)...)
}

// Info implements log.Logger.
func (l *Logger) Info(msg string, fs ...log.Field) {
	l.zap.Info(msg, l.fields(fs...)...)
}

// Warn implements log.Logger.
func (l *Logger) Warn(msg string, fs ...log.Field) {
	l.zap.Warn(msg, l.fields(fs...)...)
}

// Error implements log.Logger.
func (l *Logger) Error(msg string, fs ...log.Field) {
	l.zap.Error(msg, l.fields(fs...)...)
}

// With implements log.Logger.
func (l *Logger) With(fs ...log.Field) log.Logger {
	return &Logger{zap: l.zap.With(l.fields(fs...)...)}
}

// Sync flushes buffered logs.
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// WithContext returns a logger enriched with trace and tenant fields.
func WithContext(ctx context.Context, base log.Logger) log.Logger {
	if base == nil {
		base = &Logger{zap: zap.L()}
	}
	trace := kitctx.GetTrace(ctx)
	tenant := kitctx.GetTenant(ctx)
	fields := []log.Field{}
	if trace.TraceID != "" {
		fields = append(fields, log.F("trace_id", trace.TraceID))
	}
	if tenant.ID != "" {
		fields = append(fields, log.F("tenant_id", tenant.ID))
	}
	if user := kitctx.GetUser(ctx); user != nil {
		fields = append(fields, log.F("user_id", user.ID))
	}
	if len(fields) == 0 {
		return base
	}
	return base.With(fields...)
}

// FromContext returns a logger enriched with trace and tenant fields if present.
func FromContext(ctx context.Context, base log.Logger) log.Logger {
	return WithContext(ctx, base)
}
