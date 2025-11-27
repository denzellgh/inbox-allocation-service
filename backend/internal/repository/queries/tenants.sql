-- name: CreateTenant :exec
INSERT INTO tenants (id, name, priority_weight_alpha, priority_weight_beta, created_at, updated_at, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantByName :one
SELECT * FROM tenants WHERE name = $1;

-- name: UpdateTenant :exec
UPDATE tenants
SET name = $2,
    priority_weight_alpha = $3,
    priority_weight_beta = $4,
    updated_at = $5,
    updated_by = $6
WHERE id = $1;

-- name: DeleteTenant :exec
DELETE FROM tenants WHERE id = $1;

-- name: ListTenants :many
SELECT * FROM tenants ORDER BY created_at DESC;
