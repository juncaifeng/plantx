package repo

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/plantx/platform/iam-service/internal/domain"
)

// InMemoryRepo is an in-memory IAM repository.
type InMemoryRepo struct {
	mu          sync.RWMutex
	users       map[string]*domain.User
	roles       map[string]*domain.Role
	permissions map[string]*domain.Permission
}

// NewInMemoryRepo creates a new InMemoryRepo with default roles.
func NewInMemoryRepo() *InMemoryRepo {
	r := &InMemoryRepo{
		users:       make(map[string]*domain.User),
		roles:       make(map[string]*domain.Role),
		permissions: make(map[string]*domain.Permission),
	}
	r.roles["role_admin"] = &domain.Role{
		ID:          "role_admin",
		Name:        "Platform Admin",
		Permissions: []string{"*"},
	}
	r.roles["role_tenant_admin"] = &domain.Role{
		ID:          "role_tenant_admin",
		Name:        "Tenant Admin",
		Permissions: []string{"tenant:read", "tenant:write"},
	}
	return r
}

// CreateUser stores a new user.
func (r *InMemoryRepo) CreateUser(_ context.Context, username, tenantID string, roleIDs []string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user := &domain.User{
		ID:       uuid.NewString(),
		Username: username,
		TenantID: tenantID,
		RoleIDs:  roleIDs,
	}
	r.users[user.ID] = user
	return user, nil
}

// ListUsers returns all users.
func (r *InMemoryRepo) ListUsers(_ context.Context) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, u)
	}
	return out, nil
}

// ListRoles returns all roles.
func (r *InMemoryRepo) ListRoles(_ context.Context) ([]*domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.Role, 0, len(r.roles))
	for _, role := range r.roles {
		out = append(out, role)
	}
	return out, nil
}

// GetRole returns a role by ID.
func (r *InMemoryRepo) GetRole(_ context.Context, id string) (*domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.roles[id], nil
}

// CreateRole creates a new role.
func (r *InMemoryRepo) CreateRole(_ context.Context, name, description string, permissions []string) (*domain.Role, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	role := &domain.Role{
		ID:          uuid.NewString(),
		Name:        name,
		Description: description,
		Permissions: permissions,
	}
	r.roles[role.ID] = role
	return role, nil
}

// UpdateRole updates an existing role.
func (r *InMemoryRepo) UpdateRole(_ context.Context, id, name, description string, permissions []string) (*domain.Role, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	role, ok := r.roles[id]
	if !ok {
		return nil, fmt.Errorf("role %q not found", id)
	}
	if name != "" {
		role.Name = name
	}
	role.Description = description
	role.Permissions = permissions
	return role, nil
}

// DeleteRole removes a role by ID.
func (r *InMemoryRepo) DeleteRole(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.roles, id)
	return nil
}

// ListPermissions returns all permissions.
func (r *InMemoryRepo) ListPermissions(_ context.Context) ([]*domain.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.Permission, 0, len(r.permissions))
	for _, p := range r.permissions {
		out = append(out, p)
	}
	return out, nil
}

// CreatePermission creates a new permission.
func (r *InMemoryRepo) CreatePermission(_ context.Context, name, resource, operation, description string) (*domain.Permission, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	perm := &domain.Permission{
		ID:          uuid.NewString(),
		Name:        name,
		Resource:    resource,
		Operation:   operation,
		Description: description,
	}
	r.permissions[perm.ID] = perm
	return perm, nil
}

// DeletePermission removes a permission by ID.
func (r *InMemoryRepo) DeletePermission(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.permissions, id)
	return nil
}
