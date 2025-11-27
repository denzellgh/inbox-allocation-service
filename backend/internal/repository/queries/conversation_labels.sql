-- name: CreateConversationLabel :exec
INSERT INTO conversation_labels (id, conversation_id, label_id, created_at)
VALUES ($1, $2, $3, $4);

-- name: GetConversationLabelsByConversationID :many
SELECT * FROM conversation_labels WHERE conversation_id = $1;

-- name: GetConversationLabelsByLabelID :many
SELECT * FROM conversation_labels WHERE label_id = $1;

-- name: DeleteConversationLabel :exec
DELETE FROM conversation_labels WHERE conversation_id = $1 AND label_id = $2;

-- name: DeleteAllConversationLabels :exec
DELETE FROM conversation_labels WHERE conversation_id = $1;

-- name: CheckConversationLabelExists :one
SELECT EXISTS(
    SELECT 1 FROM conversation_labels 
    WHERE conversation_id = $1 AND label_id = $2
) AS exists;
