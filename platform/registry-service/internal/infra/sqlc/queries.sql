-- name: UpsertService :one
INSERT INTO registry_services (name, grpc_host, rest_prefix, application_id, status)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (name) DO UPDATE SET
    grpc_host = EXCLUDED.grpc_host,
    rest_prefix = EXCLUDED.rest_prefix,
    application_id = EXCLUDED.application_id,
    status = EXCLUDED.status,
    updated_at = now()
RETURNING *;

-- name: DeleteService :exec
DELETE FROM registry_services WHERE id = $1;

-- name: GetService :one
SELECT * FROM registry_services WHERE id = $1;

-- name: GetServiceByName :one
SELECT * FROM registry_services WHERE name = $1;

-- name: UpdateServiceStatusByName :one
UPDATE registry_services SET
    status = $2,
    updated_at = now()
WHERE name = $1
RETURNING *;

-- name: ListServices :many
SELECT * FROM registry_services ORDER BY created_at;

-- name: UpsertMicroApp :one
INSERT INTO micro_apps (
    service_id,
    name,
    route,
    bundle_url,
    menu_label_key,
    require_permission,
    application_id,
    upstream,
    status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (service_id, name) DO UPDATE SET
    route = EXCLUDED.route,
    bundle_url = EXCLUDED.bundle_url,
    menu_label_key = EXCLUDED.menu_label_key,
    require_permission = EXCLUDED.require_permission,
    application_id = EXCLUDED.application_id,
    upstream = EXCLUDED.upstream,
    status = EXCLUDED.status,
    updated_at = now()
RETURNING *;

-- name: ListMicroAppsByServiceID :many
SELECT * FROM micro_apps WHERE service_id = $1 ORDER BY created_at;

-- name: ListMicroApps :many
SELECT * FROM micro_apps ORDER BY created_at;

-- name: ListMicroAppsByApplicationID :many
SELECT * FROM micro_apps WHERE application_id = $1 ORDER BY created_at;

-- name: UpdateMicroApp :one
UPDATE micro_apps SET
    route = $2,
    bundle_url = $3,
    menu_label_key = $4,
    require_permission = $5,
    upstream = $6,
    status = $7,
    updated_at = now()
WHERE name = $1
RETURNING *;

-- name: DeleteMicroApp :exec
DELETE FROM micro_apps WHERE name = $1;

-- name: CreateMenu :one
INSERT INTO menus (
    label_key,
    route,
    icon,
    parent_id,
    sort_order,
    micro_app_name,
    require_permission,
    application_id,
    status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (application_id, label_key, route) DO UPDATE SET
    icon = EXCLUDED.icon,
    parent_id = EXCLUDED.parent_id,
    sort_order = EXCLUDED.sort_order,
    micro_app_name = EXCLUDED.micro_app_name,
    require_permission = EXCLUDED.require_permission,
    status = EXCLUDED.status,
    updated_at = now()
RETURNING *;

-- name: ListMenus :many
SELECT * FROM menus ORDER BY parent_id NULLS FIRST, sort_order, label_key;

-- name: ListMenusByApplicationID :many
SELECT * FROM menus WHERE application_id = $1 ORDER BY parent_id NULLS FIRST, sort_order, label_key;

-- name: GetMenu :one
SELECT * FROM menus WHERE id = $1;

-- name: UpdateMenu :one
UPDATE menus SET
    label_key = $2,
    route = $3,
    icon = $4,
    parent_id = $5,
    sort_order = $6,
    micro_app_name = $7,
    require_permission = $8,
    application_id = $9,
    status = $10,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteMenu :exec
DELETE FROM menus WHERE id = $1;

-- name: UpdateMenuSortOrder :one
UPDATE menus SET
    sort_order = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpsertRoutePolicy :one
INSERT INTO route_policies (
    service_id,
    rate_limit_rps,
    auth_required,
    canary_weight,
    canary_host
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (service_id) DO UPDATE SET
    rate_limit_rps = EXCLUDED.rate_limit_rps,
    auth_required = EXCLUDED.auth_required,
    canary_weight = EXCLUDED.canary_weight,
    canary_host = EXCLUDED.canary_host,
    updated_at = now()
RETURNING *;

-- name: GetRoutePolicy :one
SELECT * FROM route_policies WHERE service_id = $1;

-- name: ListRoutePolicies :many
SELECT * FROM route_policies ORDER BY service_id;

-- name: DeleteRoutePolicy :exec
DELETE FROM route_policies WHERE service_id = $1;

-- name: CreateApplication :one
INSERT INTO applications (
    key,
    name,
    label_key,
    icon,
    description,
    status,
    sort_order
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListApplications :many
SELECT * FROM applications ORDER BY sort_order, key;

-- name: GetApplication :one
SELECT * FROM applications WHERE id = $1;

-- name: GetApplicationByKey :one
SELECT * FROM applications WHERE key = $1;

-- name: UpdateApplication :one
UPDATE applications SET
    key = $2,
    name = $3,
    label_key = $4,
    icon = $5,
    description = $6,
    status = $7,
    sort_order = $8,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteApplication :exec
DELETE FROM applications WHERE id = $1;
