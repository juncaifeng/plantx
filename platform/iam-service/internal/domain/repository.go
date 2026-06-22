package domain

import "context"

// Repository defines persistence operations for IAM.
type Repository interface {
	CreateUser(ctx context.Context, username, tenantID string, roleIDs []string) (*User, error)
	ListUsers(ctx context.Context) ([]*User, error)

	ListRoles(ctx context.Context) ([]*Role, error)
	GetRole(ctx context.Context, id string) (*Role, error)
	CreateRole(ctx context.Context, name, description string, permissions []string) (*Role, error)
	UpdateRole(ctx context.Context, id, name, description string, permissions []string) (*Role, error)
	DeleteRole(ctx context.Context, id string) error

	ListPermissions(ctx context.Context) ([]*Permission, error)
	CreatePermission(ctx context.Context, name, resource, operation, description string) (*Permission, error)
	DeletePermission(ctx context.Context, id string) error
}
