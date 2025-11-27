-- name: CreateInbox :exec
INSERT INTO inboxes (id, tenant_id, phone_number, display_name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetInboxByID :one
SELECT * FROM inboxes WHERE id = $1;

-- name: GetInboxesByTenantID :many
SELECT * FROM inboxes WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: GetInboxByPhoneNumber :one
SELECT * FROM inboxes WHERE tenant_id = $1 AND phone_number = $2;

-- name: UpdateInbox :exec
UPDATE inboxes
SET phone_number = $2,
    display_name = $3,
    updated_at = $4
WHERE id = $1;

-- name: DeleteInbox :exec
DELETE FROM inboxes WHERE id = $1;
