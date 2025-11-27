-- name: CreateConversationRef :exec
INSERT INTO conversation_refs (
    id, tenant_id, inbox_id, external_conversation_id, customer_phone_number,
    state, assigned_operator_id, last_message_at, message_count, priority_score,
    created_at, updated_at, resolved_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: GetConversationRefByID :one
SELECT * FROM conversation_refs WHERE id = $1;

-- name: GetConversationRefByExternalID :one
SELECT * FROM conversation_refs 
WHERE tenant_id = $1 AND external_conversation_id = $2;

-- name: UpdateConversationRef :exec
UPDATE conversation_refs
SET inbox_id = $2,
    state = $3,
    assigned_operator_id = $4,
    last_message_at = $5,
    message_count = $6,
    priority_score = $7,
    updated_at = $8,
    resolved_at = $9
WHERE id = $1;

-- name: DeleteConversationRef :exec
DELETE FROM conversation_refs WHERE id = $1;

-- name: SearchConversationsByPhone :many
SELECT * FROM conversation_refs
WHERE tenant_id = $1 AND customer_phone_number = $2
ORDER BY created_at DESC;

-- name: GetConversationsByOperatorID :many
SELECT * FROM conversation_refs
WHERE tenant_id = $1 AND assigned_operator_id = $2
ORDER BY created_at DESC;

-- name: GetConversationsByOperatorAndState :many
SELECT * FROM conversation_refs
WHERE tenant_id = $1 
  AND assigned_operator_id = $2 
  AND state = $3
ORDER BY created_at DESC;

-- name: GetQueuedConversationsByTenant :many
SELECT * FROM conversation_refs
WHERE tenant_id = $1 AND state = 'QUEUED'
ORDER BY priority_score DESC, last_message_at ASC
LIMIT $2;

-- name: GetConversationsByTenantAndState :many
SELECT * FROM conversation_refs
WHERE tenant_id = $1 AND state = $2
ORDER BY created_at DESC
LIMIT $3;

-- name: GetConversationsByInbox :many
SELECT * FROM conversation_refs
WHERE tenant_id = $1 AND inbox_id = $2
ORDER BY created_at DESC
LIMIT $3;

-- CRITICAL: Allocation query with FOR UPDATE SKIP LOCKED
-- name: GetNextConversationsForAllocation :many
SELECT * FROM conversation_refs
WHERE tenant_id = $1 
  AND inbox_id = ANY($2::uuid[])
  AND state = 'QUEUED'
ORDER BY priority_score DESC, last_message_at ASC
LIMIT $3
FOR UPDATE SKIP LOCKED;

-- CRITICAL: Lock specific conversation for claim
-- name: LockConversationForClaim :one
SELECT * FROM conversation_refs
WHERE id = $1 AND state = 'QUEUED'
FOR UPDATE NOWAIT;

-- Update state only (for allocation/deallocate/resolve)
-- name: UpdateConversationState :exec
UPDATE conversation_refs
SET state = $2,
    assigned_operator_id = $3,
    updated_at = $4,
    resolved_at = $5
WHERE id = $1;
