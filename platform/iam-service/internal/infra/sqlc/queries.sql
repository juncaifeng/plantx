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

-- name: CreateAttribute :one
INSERT INTO attributes (key, value_type, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListAttributes :many
SELECT * FROM attributes
ORDER BY key;

-- name: UpdateAttribute :one
UPDATE attributes
SET key = $2,
    value_type = $3,
    description = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteAttribute :exec
DELETE FROM attributes WHERE id = $1;

-- name: CreateCondition :one
INSERT INTO conditions (name, attribute_key, operator, value, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListConditions :many
SELECT * FROM conditions
ORDER BY name;

-- name: UpdateCondition :one
UPDATE conditions
SET name = $2,
    attribute_key = $3,
    operator = $4,
    value = $5,
    description = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteCondition :exec
DELETE FROM conditions WHERE id = $1;

-- name: CreatePolicy :one
INSERT INTO policies (name, description, effect, priority)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListPolicies :many
SELECT * FROM policies
ORDER BY priority DESC, name;

-- name: UpdatePolicy :one
UPDATE policies
SET name = $2,
    description = $3,
    effect = $4,
    priority = $5,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeletePolicy :exec
DELETE FROM policies WHERE id = $1;

-- name: AddPolicyPermission :exec
INSERT INTO policy_permissions (policy_id, permission)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemovePolicyPermissions :exec
DELETE FROM policy_permissions WHERE policy_id = $1;

-- name: ListPolicyPermissions :many
SELECT permission FROM policy_permissions WHERE policy_id = $1;

-- name: AddPolicyCondition :exec
INSERT INTO policy_conditions (policy_id, condition_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemovePolicyConditions :exec
DELETE FROM policy_conditions WHERE policy_id = $1;

-- name: ListPolicyConditions :many
SELECT condition_id FROM policy_conditions WHERE policy_id = $1;
