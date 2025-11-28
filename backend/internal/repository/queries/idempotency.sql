-- name: CreateIdempotencyKey :exec
INSERT INTO idempotency_keys (
    id, key, tenant_id, endpoint, method, request_hash,
    response_status, response_body, created_at, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetIdempotencyKey :one
SELECT * FROM idempotency_keys
WHERE tenant_id = $1 AND key = $2;

-- name: DeleteIdempotencyKey :exec
DELETE FROM idempotency_keys WHERE id = $1;

-- name: DeleteExpiredIdempotencyKeys :execrows
DELETE FROM idempotency_keys
WHERE expires_at < NOW();

-- name: GetExpiredIdempotencyKeysForCleanup :many
SELECT * FROM idempotency_keys
WHERE expires_at < NOW()
ORDER BY expires_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: CountIdempotencyKeys :one
SELECT COUNT(*) FROM idempotency_keys WHERE tenant_id = $1;
