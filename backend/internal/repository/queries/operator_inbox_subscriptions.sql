-- name: CreateSubscription :exec
INSERT INTO operator_inbox_subscriptions (id, operator_id, inbox_id, created_at)
VALUES ($1, $2, $3, $4);

-- name: GetSubscriptionByID :one
SELECT * FROM operator_inbox_subscriptions WHERE id = $1;

-- name: GetSubscriptionsByOperatorID :many
SELECT * FROM operator_inbox_subscriptions WHERE operator_id = $1;

-- name: GetSubscriptionsByInboxID :many
SELECT * FROM operator_inbox_subscriptions WHERE inbox_id = $1;

-- name: GetSubscriptionByOperatorAndInbox :one
SELECT * FROM operator_inbox_subscriptions 
WHERE operator_id = $1 AND inbox_id = $2;

-- name: DeleteSubscription :exec
DELETE FROM operator_inbox_subscriptions WHERE id = $1;

-- name: DeleteSubscriptionByOperatorAndInbox :exec
DELETE FROM operator_inbox_subscriptions 
WHERE operator_id = $1 AND inbox_id = $2;

-- name: GetSubscribedInboxIDs :many
SELECT inbox_id FROM operator_inbox_subscriptions WHERE operator_id = $1;

-- name: CheckSubscriptionExists :one
SELECT EXISTS(
    SELECT 1 FROM operator_inbox_subscriptions 
    WHERE operator_id = $1 AND inbox_id = $2
) AS exists;
