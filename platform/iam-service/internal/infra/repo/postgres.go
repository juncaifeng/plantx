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
