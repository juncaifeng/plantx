package app

import (
	"context"

	"github.com/plantx/platform/iam-service/internal/domain"
)

// IAMService implements IAM use cases.
type IAMService struct {
	repo domain.Repository
}

// NewIAMService creates a new IAMService.
func NewIAMService(repo domain.Repository) *IAMService {
	return &IAMService{repo: repo}
}

// CreateUser creates a new user.
func (s *IAMService) CreateUser(ctx context.Context, username, tenantID string, roleIDs []string) (*domain.User, error) {
	return s.repo.CreateUser(ctx, username, tenantID, roleIDs)
}

// ListUsers lists all users.
func (s *IAMService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return s.repo.ListUsers(ctx)
}

// ListRoles lists all roles.
func (s *IAMService) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.repo.ListRoles(ctx)
}

// GetRole returns a role by ID.
func (s *IAMService) GetRole(ctx context.Context, id string) (*domain.Role, error) {
	return s.repo.GetRole(ctx, id)
}

// CreateRole creates a new role.
func (s *IAMService) CreateRole(ctx context.Context, name, description string, permissions []string) (*domain.Role, error) {
	return s.repo.CreateRole(ctx, name, description, permissions)
}

// UpdateRole updates an existing role.
func (s *IAMService) UpdateRole(ctx context.Context, id, name, description string, permissions []string) (*domain.Role, error) {
	return s.repo.UpdateRole(ctx, id, name, description, permissions)
}

// DeleteRole deletes a role by ID.
func (s *IAMService) DeleteRole(ctx context.Context, id string) error {
	return s.repo.DeleteRole(ctx, id)
}

// ListPermissions lists all permissions.
func (s *IAMService) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

// CreatePermission creates a new permission.
func (s *IAMService) CreatePermission(ctx context.Context, name, resource, operation, description string) (*domain.Permission, error) {
	return s.repo.CreatePermission(ctx, name, resource, operation, description)
}

// DeletePermission deletes a permission by ID.
func (s *IAMService) DeletePermission(ctx context.Context, id string) error {
	return s.repo.DeletePermission(ctx, id)
}

// ListAttributes lists all ABAC attributes.
func (s *IAMService) ListAttributes(ctx context.Context) ([]*domain.Attribute, error) {
	return s.repo.ListAttributes(ctx)
}

// CreateAttribute creates a new ABAC attribute.
func (s *IAMService) CreateAttribute(ctx context.Context, key, valueType, description string) (*domain.Attribute, error) {
	return s.repo.CreateAttribute(ctx, key, valueType, description)
}

// UpdateAttribute updates an existing ABAC attribute.
func (s *IAMService) UpdateAttribute(ctx context.Context, id, key, valueType, description string) (*domain.Attribute, error) {
	return s.repo.UpdateAttribute(ctx, id, key, valueType, description)
}

// DeleteAttribute deletes an ABAC attribute by ID.
func (s *IAMService) DeleteAttribute(ctx context.Context, id string) error {
	return s.repo.DeleteAttribute(ctx, id)
}

// ListConditions lists all ABAC conditions.
func (s *IAMService) ListConditions(ctx context.Context) ([]*domain.Condition, error) {
	return s.repo.ListConditions(ctx)
}

// CreateCondition creates a new ABAC condition.
func (s *IAMService) CreateCondition(ctx context.Context, name, attributeKey, operator, value, description string) (*domain.Condition, error) {
	return s.repo.CreateCondition(ctx, name, attributeKey, operator, value, description)
}

// UpdateCondition updates an existing ABAC condition.
func (s *IAMService) UpdateCondition(ctx context.Context, id, name, attributeKey, operator, value, description string) (*domain.Condition, error) {
	return s.repo.UpdateCondition(ctx, id, name, attributeKey, operator, value, description)
}

// DeleteCondition deletes an ABAC condition by ID.
func (s *IAMService) DeleteCondition(ctx context.Context, id string) error {
	return s.repo.DeleteCondition(ctx, id)
}

// ListPolicies lists all ABAC policies.
func (s *IAMService) ListPolicies(ctx context.Context) ([]*domain.Policy, error) {
	return s.repo.ListPolicies(ctx)
}

// CreatePolicy creates a new ABAC policy.
func (s *IAMService) CreatePolicy(ctx context.Context, name, description, effect string, priority int32, permissions, conditionIDs []string) (*domain.Policy, error) {
	return s.repo.CreatePolicy(ctx, name, description, effect, priority, permissions, conditionIDs)
}

// UpdatePolicy updates an existing ABAC policy.
func (s *IAMService) UpdatePolicy(ctx context.Context, id, name, description, effect string, priority int32, permissions, conditionIDs []string) (*domain.Policy, error) {
	return s.repo.UpdatePolicy(ctx, id, name, description, effect, priority, permissions, conditionIDs)
}

// DeletePolicy deletes an ABAC policy by ID.
func (s *IAMService) DeletePolicy(ctx context.Context, id string) error {
	return s.repo.DeletePolicy(ctx, id)
}
