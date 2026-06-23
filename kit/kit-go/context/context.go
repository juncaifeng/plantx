// Package kitctx provides request-scoped context helpers for user, tenant, and trace data.
package kitctx

import (
	"context"

	"github.com/plantx/kit/kit-go/auth"
	"github.com/plantx/kit/kit-go/tenant"
)

type ctxKey int

const (
	userKey ctxKey = iota
	tenantKey
	traceKey
)

// WithUser injects an authenticated user into the context.
func WithUser(ctx context.Context, u *auth.UserInfo) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// GetUser returns the authenticated user from the context, or nil.
func GetUser(ctx context.Context) *auth.UserInfo {
	if v, ok := ctx.Value(userKey).(*auth.UserInfo); ok {
		return v
	}
	return nil
}

// WithTenant injects tenant info into the context.
func WithTenant(ctx context.Context, t tenant.Info) context.Context {
	return context.WithValue(ctx, tenantKey, t)
}

// GetTenant returns tenant info from the context.
func GetTenant(ctx context.Context) tenant.Info {
	if v, ok := ctx.Value(tenantKey).(tenant.Info); ok {
		return v
	}
	return tenant.Info{}
}

// TraceContext carries distributed tracing identifiers.
type TraceContext struct {
	TraceID string
	SpanID  string
}

// WithTrace injects trace context.
func WithTrace(ctx context.Context, tc TraceContext) context.Context {
	return context.WithValue(ctx, traceKey, tc)
}

// GetTrace returns trace context from the context.
func GetTrace(ctx context.Context) TraceContext {
	if v, ok := ctx.Value(traceKey).(TraceContext); ok {
		return v
	}
	return TraceContext{}
}
