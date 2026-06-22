-- name: CreateUser :one
INSERT INTO users (username, tenant_id, role_ids)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at;

-- name: CreatePermission :one
INSERT INTO permissions (name, resource, operation, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListPermissions :many
SELECT * FROM permissions
ORDER BY name;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1;

-- name: CreateRole :one
INSERT INTO roles (name, description, permissions)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles
SET name = $2,
    description = $3,
    permissions = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;

-- name: GetRole :one
SELECT * FROM roles WHERE id = $1;

-- name: ListRoles :many
SELECT * FROM roles
ORDER BY name;
