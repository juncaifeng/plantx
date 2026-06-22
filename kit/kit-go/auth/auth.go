// Package auth defines the authentication abstraction and user identity model.
package auth

import "context"

// UserInfo represents an authenticated user and their platform-level claims.
type UserInfo struct {
	ID          string
	TenantID    string
	Username    string
	DisplayName string
	Email       string
	Roles       []string
	Permissions []string
	Claims      map[string]string
}

// Authenticator resolves a raw credential (e.g. bearer token) into UserInfo.
type Authenticator interface {
	Authenticate(ctx context.Context, credential string) (*UserInfo, error)
}
