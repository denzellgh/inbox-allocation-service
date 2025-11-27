-- name: CreateOperatorStatus :exec
INSERT INTO operator_status (id, operator_id, status, last_status_change_at)
VALUES ($1, $2, $3, $4);

-- name: GetOperatorStatusByOperatorID :one
SELECT * FROM operator_status WHERE operator_id = $1;

-- name: UpdateOperatorStatus :exec
UPDATE operator_status
SET status = $2,
    last_status_change_at = $3
WHERE operator_id = $1;

-- name: GetAvailableOperators :many
SELECT os.*
FROM operator_status os
JOIN operators o ON o.id = os.operator_id
WHERE o.tenant_id = $1 AND os.status = 'AVAILABLE';
