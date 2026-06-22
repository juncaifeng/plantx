-- name: CreateOrder :one
INSERT INTO orders (tenant_id, customer_name, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOrder :one
SELECT * FROM orders
WHERE id = $1 AND tenant_id = $2;

-- name: ListOrdersByStatus :many
SELECT * FROM orders
WHERE tenant_id = $1 AND status = $2
ORDER BY created_at DESC;
