package authz

import (
	"context"

	"github.com/plantx/kit/kit-go/auth"
)

// Action identifies a resource operation in a policy decision.
type Action struct {
	Service   string
	Resource  string
	Operation string
}

// Request carries all inputs for an authorization decision.
type Request struct {
	User            auth.UserInfo
	Action          Action
	ResourceContext map[string]any
}

// Decision is the result of an authorization check.
type Decision struct {
	Allowed bool
	Reason  string
}

// Authorizer decides whether a request is allowed.
type Authorizer interface {
	Authorize(ctx context.Context, req Request) (Decision, error)
}
