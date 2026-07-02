package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/plantx/kit/kit-go/db"
	"github.com/plantx/platform/iam-service/internal/domain"
	"github.com/plantx/platform/iam-service/internal/infra/sqlc"
)

// PostgresRepo implements domain.Repository using PostgreSQL via sqlc.
type PostgresRepo struct {
	queries *sqlc.Queries
}

// NewPostgresRepo creates a new PostgresRepo.
func NewPostgresRepo(d db.DB) *PostgresRepo {
	return &PostgresRepo{queries: sqlc.New(d)}
}

// CreateUser persists a new user.
func (r *PostgresRepo) CreateUser(ctx context.Context, username, tenantID string, roleIDs []string) (*domain.User, error) {
	row, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Username: username,
		TenantID: tenantID,
		RoleIds:  roleIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return toDomainUser(row), nil
}

// ListUsers returns all users.
func (r *PostgresRepo) ListUsers(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	out := make([]*domain.User, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainUser(row))
	}
	return out, nil
}

// ListPermissions returns all permissions.
func (r *PostgresRepo) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	rows, err := r.queries.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	out := make([]*domain.Permission, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainPermission(row))
	}
	return out, nil
}

// CreatePermission persists a new permission.
func (r *PostgresRepo) CreatePermission(ctx context.Context, name, resource, operation, description string) (*domain.Permission, error) {
	row, err := r.queries.CreatePermission(ctx, sqlc.CreatePermissionParams{
		Name:        name,
		Resource:    resource,
		Operation:   operation,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("create permission: %w", err)
	}
	return toDomainPermission(row), nil
}

// DeletePermission removes a permission by ID.
func (r *PostgresRepo) DeletePermission(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid permission id: %w", err)
	}
	if err := r.queries.DeletePermission(ctx, uid); err != nil {
		return fmt.Errorf("delete permission: %w", err)
	}
	return nil
}

// ListRoles returns all roles.
func (r *PostgresRepo) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	rows, err := r.queries.ListRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	out := make([]*domain.Role, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainRole(row))
	}
	return out, nil
}

// GetRole returns a role by ID.
func (r *PostgresRepo) GetRole(ctx context.Context, id string) (*domain.Role, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid role id: %w", err)
	}
	row, err := r.queries.GetRole(ctx, uid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}
	return toDomainRole(row), nil
}

// CreateRole persists a new role.
func (r *PostgresRepo) CreateRole(ctx context.Context, name, description string, permissions []string) (*domain.Role, error) {
	row, err := r.queries.CreateRole(ctx, sqlc.CreateRoleParams{
		Name:        name,
		Description: description,
		Permissions: permissions,
	})
	if err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}
	return toDomainRole(row), nil
}

// UpdateRole updates an existing role.
func (r *PostgresRepo) UpdateRole(ctx context.Context, id, name, description string, permissions []string) (*domain.Role, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid role id: %w", err)
	}
	row, err := r.queries.UpdateRole(ctx, sqlc.UpdateRoleParams{
		ID:          uid,
		Name:        name,
		Description: description,
		Permissions: permissions,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("role %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}
	return toDomainRole(row), nil
}

// DeleteRole removes a role by ID.
func (r *PostgresRepo) DeleteRole(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid role id: %w", err)
	}
	if err := r.queries.DeleteRole(ctx, uid); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

// ListAttributes returns all ABAC attributes.
func (r *PostgresRepo) ListAttributes(ctx context.Context) ([]*domain.Attribute, error) {
	rows, err := r.queries.ListAttributes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list attributes: %w", err)
	}
	out := make([]*domain.Attribute, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainAttribute(row))
	}
	return out, nil
}

// CreateAttribute persists a new ABAC attribute.
func (r *PostgresRepo) CreateAttribute(ctx context.Context, key, valueType, description string) (*domain.Attribute, error) {
	row, err := r.queries.CreateAttribute(ctx, sqlc.CreateAttributeParams{
		Key:         key,
		ValueType:   valueType,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("create attribute: %w", err)
	}
	return toDomainAttribute(row), nil
}

// UpdateAttribute updates an existing ABAC attribute.
func (r *PostgresRepo) UpdateAttribute(ctx context.Context, id, key, valueType, description string) (*domain.Attribute, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute id: %w", err)
	}
	row, err := r.queries.UpdateAttribute(ctx, sqlc.UpdateAttributeParams{
		ID:          uid,
		Key:         key,
		ValueType:   valueType,
		Description: description,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("attribute %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("update attribute: %w", err)
	}
	return toDomainAttribute(row), nil
}

// DeleteAttribute removes an ABAC attribute by ID.
func (r *PostgresRepo) DeleteAttribute(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid attribute id: %w", err)
	}
	if err := r.queries.DeleteAttribute(ctx, uid); err != nil {
		return fmt.Errorf("delete attribute: %w", err)
	}
	return nil
}

// ListConditions returns all ABAC conditions.
func (r *PostgresRepo) ListConditions(ctx context.Context) ([]*domain.Condition, error) {
	rows, err := r.queries.ListConditions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list conditions: %w", err)
	}
	out := make([]*domain.Condition, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainCondition(row))
	}
	return out, nil
}

// CreateCondition persists a new ABAC condition.
func (r *PostgresRepo) CreateCondition(ctx context.Context, name, attributeKey, operator, value, description string) (*domain.Condition, error) {
	row, err := r.queries.CreateCondition(ctx, sqlc.CreateConditionParams{
		Name:         name,
		AttributeKey: attributeKey,
		Operator:     operator,
		Value:        value,
		Description:  description,
	})
	if err != nil {
		return nil, fmt.Errorf("create condition: %w", err)
	}
	return toDomainCondition(row), nil
}

// UpdateCondition updates an existing ABAC condition.
func (r *PostgresRepo) UpdateCondition(ctx context.Context, id, name, attributeKey, operator, value, description string) (*domain.Condition, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid condition id: %w", err)
	}
	row, err := r.queries.UpdateCondition(ctx, sqlc.UpdateConditionParams{
		ID:           uid,
		Name:         name,
		AttributeKey: attributeKey,
		Operator:     operator,
		Value:        value,
		Description:  description,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("condition %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("update condition: %w", err)
	}
	return toDomainCondition(row), nil
}

// DeleteCondition removes an ABAC condition by ID.
func (r *PostgresRepo) DeleteCondition(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid condition id: %w", err)
	}
	if err := r.queries.DeleteCondition(ctx, uid); err != nil {
		return fmt.Errorf("delete condition: %w", err)
	}
	return nil
}

// ListPolicies returns all ABAC policies with their permissions and condition IDs.
func (r *PostgresRepo) ListPolicies(ctx context.Context) ([]*domain.Policy, error) {
	rows, err := r.queries.ListPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	out := make([]*domain.Policy, 0, len(rows))
	for _, row := range rows {
		p := toDomainPolicy(row)
		perms, err := r.queries.ListPolicyPermissions(ctx, row.ID)
		if err != nil {
			return nil, fmt.Errorf("list policy permissions: %w", err)
		}
		for _, perm := range perms {
			p.Permissions = append(p.Permissions, perm)
		}
		conds, err := r.queries.ListPolicyConditions(ctx, row.ID)
		if err != nil {
			return nil, fmt.Errorf("list policy conditions: %w", err)
		}
		for _, cond := range conds {
			p.ConditionIDs = append(p.ConditionIDs, cond.String())
		}
		out = append(out, p)
	}
	return out, nil
}

// CreatePolicy persists a new ABAC policy.
func (r *PostgresRepo) CreatePolicy(ctx context.Context, name, description, effect string, priority int32, permissions, conditionIDs []string) (*domain.Policy, error) {
	row, err := r.queries.CreatePolicy(ctx, sqlc.CreatePolicyParams{
		Name:        name,
		Description: description,
		Effect:      effect,
		Priority:    priority,
	})
	if err != nil {
		return nil, fmt.Errorf("create policy: %w", err)
	}
	if err := r.setPolicyRelations(ctx, row.ID, permissions, conditionIDs); err != nil {
		return nil, err
	}
	return r.policyWithRelations(ctx, row.ID)
}

// UpdatePolicy updates an existing ABAC policy.
func (r *PostgresRepo) UpdatePolicy(ctx context.Context, id, name, description, effect string, priority int32, permissions, conditionIDs []string) (*domain.Policy, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid policy id: %w", err)
	}
	row, err := r.queries.UpdatePolicy(ctx, sqlc.UpdatePolicyParams{
		ID:          uid,
		Name:        name,
		Description: description,
		Effect:      effect,
		Priority:    priority,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("policy %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("update policy: %w", err)
	}
	if err := r.queries.RemovePolicyPermissions(ctx, row.ID); err != nil {
		return nil, fmt.Errorf("remove policy permissions: %w", err)
	}
	if err := r.queries.RemovePolicyConditions(ctx, row.ID); err != nil {
		return nil, fmt.Errorf("remove policy conditions: %w", err)
	}
	if err := r.setPolicyRelations(ctx, row.ID, permissions, conditionIDs); err != nil {
		return nil, err
	}
	return r.policyWithRelations(ctx, row.ID)
}

// DeletePolicy removes an ABAC policy by ID.
func (r *PostgresRepo) DeletePolicy(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid policy id: %w", err)
	}
	if err := r.queries.DeletePolicy(ctx, uid); err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}
	return nil
}

func (r *PostgresRepo) setPolicyRelations(ctx context.Context, policyID uuid.UUID, permissions, conditionIDs []string) error {
	for _, perm := range permissions {
		if err := r.queries.AddPolicyPermission(ctx, sqlc.AddPolicyPermissionParams{
			PolicyID:   policyID,
			Permission: perm,
		}); err != nil {
			return fmt.Errorf("add policy permission: %w", err)
		}
	}
	for _, condID := range conditionIDs {
		cuid, err := uuid.Parse(condID)
		if err != nil {
			return fmt.Errorf("invalid condition id %q: %w", condID, err)
		}
		if err := r.queries.AddPolicyCondition(ctx, sqlc.AddPolicyConditionParams{
			PolicyID:     policyID,
			ConditionID: cuid,
		}); err != nil {
			return fmt.Errorf("add policy condition: %w", err)
		}
	}
	return nil
}

func (r *PostgresRepo) policyWithRelations(ctx context.Context, policyID uuid.UUID) (*domain.Policy, error) {
	row, err := r.queries.ListPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	for _, p := range row {
		if p.ID == policyID {
			policy := toDomainPolicy(p)
			perms, err := r.queries.ListPolicyPermissions(ctx, policyID)
			if err != nil {
				return nil, fmt.Errorf("list policy permissions: %w", err)
			}
			for _, perm := range perms {
				policy.Permissions = append(policy.Permissions, perm)
			}
			conds, err := r.queries.ListPolicyConditions(ctx, policyID)
			if err != nil {
				return nil, fmt.Errorf("list policy conditions: %w", err)
			}
			for _, cond := range conds {
				policy.ConditionIDs = append(policy.ConditionIDs, cond.String())
			}
			return policy, nil
		}
	}
	return nil, fmt.Errorf("policy %q not found", policyID)
}

func toDomainUser(row sqlc.User) *domain.User {
	return &domain.User{
		ID:       row.ID.String(),
		Username: row.Username,
		TenantID: row.TenantID,
		RoleIDs:  row.RoleIds,
	}
}

func toDomainPermission(row sqlc.Permission) *domain.Permission {
	return &domain.Permission{
		ID:          row.ID.String(),
		Name:        row.Name,
		Resource:    row.Resource,
		Operation:   row.Operation,
		Description: row.Description,
	}
}

func toDomainRole(row sqlc.Role) *domain.Role {
	return &domain.Role{
		ID:          row.ID.String(),
		Name:        row.Name,
		Description: row.Description,
		Permissions: row.Permissions,
	}
}

func toDomainAttribute(row sqlc.Attribute) *domain.Attribute {
	return &domain.Attribute{
		ID:          row.ID.String(),
		Key:         row.Key,
		ValueType:   row.ValueType,
		Description: row.Description,
	}
}

func toDomainCondition(row sqlc.Condition) *domain.Condition {
	return &domain.Condition{
		ID:           row.ID.String(),
		Name:         row.Name,
		AttributeKey: row.AttributeKey,
		Operator:     row.Operator,
		Value:        row.Value,
		Description:  row.Description,
	}
}

func toDomainPolicy(row sqlc.Policy) *domain.Policy {
	return &domain.Policy{
		ID:          row.ID.String(),
		Name:        row.Name,
		Description: row.Description,
		Effect:      row.Effect,
		Priority:    row.Priority,
	}
}
