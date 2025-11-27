-- name: CreateOperator :exec
INSERT INTO operators (id, tenant_id, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetOperatorByID :one
SELECT * FROM operators WHERE id = $1;

-- name: GetOperatorsByTenantID :many
SELECT * FROM operators WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: GetOperatorsByTenantAndRole :many
SELECT * FROM operators WHERE tenant_id = $1 AND role = $2 ORDER BY created_at DESC;

-- name: UpdateOperator :exec
UPDATE operators
SET role = $2,
    updated_at = $3
WHERE id = $1;

-- name: DeleteOperator :exec
DELETE FROM operators WHERE id = $1;
