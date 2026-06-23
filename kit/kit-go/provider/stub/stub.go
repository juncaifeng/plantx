// Package stub provides test-friendly stub implementations of Kit abstractions.
package stub

import (
	"context"
	"errors"

	"github.com/plantx/kit/kit-go/auth"
	"github.com/plantx/kit/kit-go/authz"
	kitctx "github.com/plantx/kit/kit-go/context"
	"github.com/plantx/kit/kit-go/tenant"
)

// Authenticator is a stub Authenticator that validates a static set of tokens.
type Authenticator struct {
	Tokens map[string]*auth.UserInfo
}

// Authenticate validates the credential against the configured static tokens.
func (a *Authenticator) Authenticate(_ context.Context, credential string) (*auth.UserInfo, error) {
	if a == nil || a.Tokens == nil {
		return nil, errors.New("no tokens configured")
	}
	u, ok := a.Tokens[credential]
	if !ok {
		return nil, errors.New("invalid token")
	}
	return u, nil
}

// Authorizer is a stub Authorizer that allows actions matching the predicate.
type Authorizer struct {
	AllowFunc func(ctx context.Context, req authz.Request) bool
}

// Authorize evaluates the request using the configured AllowFunc predicate.
func (a *Authorizer) Authorize(ctx context.Context, req authz.Request) (authz.Decision, error) {
	if a != nil && a.AllowFunc != nil && a.AllowFunc(ctx, req) {
		return authz.Decision{Allowed: true}, nil
	}
	return authz.Decision{Allowed: false, Reason: "stub denied"}, nil
}

// AllowAll returns an Authorizer that permits every request.
func AllowAll() authz.Authorizer {
	return &Authorizer{AllowFunc: func(context.Context, authz.Request) bool { return true }}
}

// DenyAll returns an Authorizer that denies every request.
func DenyAll() authz.Authorizer {
	return &Authorizer{AllowFunc: func(context.Context, authz.Request) bool { return false }}
}

// TenantResolver resolves tenant from the tenant_id claim.
type TenantResolver struct{}

// Resolve extracts tenant information from the provided claims.
func (TenantResolver) Resolve(_ string, claims map[string]string) (tenant.Info, error) {
	if claims == nil {
		return tenant.Info{}, nil
	}
	return tenant.Info{ID: claims["tenant_id"], Name: claims["tenant_name"]}, nil
}

// WithUser injects a user into the context and derives tenant if possible.
func WithUser(ctx context.Context, u *auth.UserInfo) context.Context {
	ctx = kitctx.WithUser(ctx, u)
	if u != nil && u.TenantID != "" {
		ctx = kitctx.WithTenant(ctx, tenant.Info{ID: u.TenantID})
	}
	return ctx
}
