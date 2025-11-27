-- name: CreateGracePeriodAssignment :exec
INSERT INTO grace_period_assignments (id, conversation_id, operator_id, expires_at, reason, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetGracePeriodByConversationID :one
SELECT * FROM grace_period_assignments WHERE conversation_id = $1;

-- name: GetGracePeriodsByOperatorID :many
SELECT * FROM grace_period_assignments WHERE operator_id = $1;

-- name: GetExpiredGracePeriods :many
SELECT * FROM grace_period_assignments
WHERE expires_at <= NOW()
ORDER BY expires_at ASC
LIMIT $1;

-- CRITICAL: Get and lock expired for worker
-- name: GetAndLockExpiredGracePeriods :many
SELECT * FROM grace_period_assignments
WHERE expires_at <= NOW()
ORDER BY expires_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: DeleteGracePeriodAssignment :exec
DELETE FROM grace_period_assignments WHERE id = $1;

-- name: DeleteGracePeriodsByOperatorID :exec
DELETE FROM grace_period_assignments WHERE operator_id = $1;

-- name: DeleteGracePeriodByConversationID :exec
DELETE FROM grace_period_assignments WHERE conversation_id = $1;
