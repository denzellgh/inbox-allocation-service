-- name: CreateLabel :exec
INSERT INTO labels (id, tenant_id, inbox_id, name, color, created_by, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetLabelByID :one
SELECT * FROM labels WHERE id = $1;

-- name: GetLabelsByInboxID :many
SELECT * FROM labels WHERE tenant_id = $1 AND inbox_id = $2 ORDER BY name;

-- name: GetLabelByName :one
SELECT * FROM labels WHERE inbox_id = $1 AND name = $2;

-- name: UpdateLabel :exec
UPDATE labels
SET name = $2,
    color = $3
WHERE id = $1;

-- name: DeleteLabel :exec
DELETE FROM labels WHERE id = $1;
