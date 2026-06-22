// Package telemetry provides OpenTelemetry tracing helpers for Kit servers.
package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// InitTracerProvider configures a global tracer provider. If no OTLP endpoint is
// configured it falls back to a stdout exporter for local development.
func InitTracerProvider(ctx context.Context, serviceName string) (*sdktrace.TracerProvider, error) {
	if ep := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); ep != "" {
		// OTLP exporter can be added here when the dependency is desired.
		// For now fall back to stdout so that the trace pipeline is wired.
		_ = ep
	}
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

// Tracer returns the global tracer for the given scope.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
