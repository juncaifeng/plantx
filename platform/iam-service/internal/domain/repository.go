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

	ListAttributes(ctx context.Context) ([]*Attribute, error)
	CreateAttribute(ctx context.Context, key, valueType, description string) (*Attribute, error)
	UpdateAttribute(ctx context.Context, id, key, valueType, description string) (*Attribute, error)
	DeleteAttribute(ctx context.Context, id string) error

	ListConditions(ctx context.Context) ([]*Condition, error)
	CreateCondition(ctx context.Context, name, attributeKey, operator, value, description string) (*Condition, error)
	UpdateCondition(ctx context.Context, id, name, attributeKey, operator, value, description string) (*Condition, error)
	DeleteCondition(ctx context.Context, id string) error

	ListPolicies(ctx context.Context) ([]*Policy, error)
	CreatePolicy(ctx context.Context, name, description, effect string, priority int32, permissions, conditionIDs []string) (*Policy, error)
	UpdatePolicy(ctx context.Context, id, name, description, effect string, priority int32, permissions, conditionIDs []string) (*Policy, error)
	DeletePolicy(ctx context.Context, id string) error
}
