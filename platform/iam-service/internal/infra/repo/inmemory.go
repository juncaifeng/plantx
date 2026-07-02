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
	attributes  map[string]*domain.Attribute
	conditions  map[string]*domain.Condition
	policies    map[string]*domain.Policy
}

// NewInMemoryRepo creates a new InMemoryRepo with default roles.
func NewInMemoryRepo() *InMemoryRepo {
	r := &InMemoryRepo{
		users:       make(map[string]*domain.User),
		roles:       make(map[string]*domain.Role),
		permissions: make(map[string]*domain.Permission),
		attributes:  make(map[string]*domain.Attribute),
		conditions:  make(map[string]*domain.Condition),
		policies:    make(map[string]*domain.Policy),
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

// ListAttributes returns all ABAC attributes.
func (r *InMemoryRepo) ListAttributes(_ context.Context) ([]*domain.Attribute, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Attribute, 0, len(r.attributes))
	for _, a := range r.attributes {
		out = append(out, a)
	}
	return out, nil
}

// CreateAttribute creates a new ABAC attribute.
func (r *InMemoryRepo) CreateAttribute(_ context.Context, key, valueType, description string) (*domain.Attribute, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a := &domain.Attribute{ID: uuid.NewString(), Key: key, ValueType: valueType, Description: description}
	r.attributes[a.ID] = a
	return a, nil
}

// UpdateAttribute updates an existing ABAC attribute.
func (r *InMemoryRepo) UpdateAttribute(_ context.Context, id, key, valueType, description string) (*domain.Attribute, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.attributes[id]
	if !ok {
		return nil, fmt.Errorf("attribute %q not found", id)
	}
	a.Key = key
	a.ValueType = valueType
	a.Description = description
	return a, nil
}

// DeleteAttribute removes an ABAC attribute by ID.
func (r *InMemoryRepo) DeleteAttribute(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.attributes, id)
	return nil
}

// ListConditions returns all ABAC conditions.
func (r *InMemoryRepo) ListConditions(_ context.Context) ([]*domain.Condition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Condition, 0, len(r.conditions))
	for _, c := range r.conditions {
		out = append(out, c)
	}
	return out, nil
}

// CreateCondition creates a new ABAC condition.
func (r *InMemoryRepo) CreateCondition(_ context.Context, name, attributeKey, operator, value, description string) (*domain.Condition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c := &domain.Condition{ID: uuid.NewString(), Name: name, AttributeKey: attributeKey, Operator: operator, Value: value, Description: description}
	r.conditions[c.ID] = c
	return c, nil
}

// UpdateCondition updates an existing ABAC condition.
func (r *InMemoryRepo) UpdateCondition(_ context.Context, id, name, attributeKey, operator, value, description string) (*domain.Condition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.conditions[id]
	if !ok {
		return nil, fmt.Errorf("condition %q not found", id)
	}
	c.Name = name
	c.AttributeKey = attributeKey
	c.Operator = operator
	c.Value = value
	c.Description = description
	return c, nil
}

// DeleteCondition removes an ABAC condition by ID.
func (r *InMemoryRepo) DeleteCondition(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conditions, id)
	return nil
}

// ListPolicies returns all ABAC policies.
func (r *InMemoryRepo) ListPolicies(_ context.Context) ([]*domain.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Policy, 0, len(r.policies))
	for _, p := range r.policies {
		out = append(out, p)
	}
	return out, nil
}

// CreatePolicy creates a new ABAC policy.
func (r *InMemoryRepo) CreatePolicy(_ context.Context, name, description, effect string, priority int32, permissions, conditionIDs []string) (*domain.Policy, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p := &domain.Policy{ID: uuid.NewString(), Name: name, Description: description, Effect: effect, Priority: priority, Permissions: permissions, ConditionIDs: conditionIDs}
	r.policies[p.ID] = p
	return p, nil
}

// UpdatePolicy updates an existing ABAC policy.
func (r *InMemoryRepo) UpdatePolicy(_ context.Context, id, name, description, effect string, priority int32, permissions, conditionIDs []string) (*domain.Policy, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.policies[id]
	if !ok {
		return nil, fmt.Errorf("policy %q not found", id)
	}
	p.Name = name
	p.Description = description
	p.Effect = effect
	p.Priority = priority
	p.Permissions = permissions
	p.ConditionIDs = conditionIDs
	return p, nil
}

// DeletePolicy removes an ABAC policy by ID.
func (r *InMemoryRepo) DeletePolicy(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.policies, id)
	return nil
}
