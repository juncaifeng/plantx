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
