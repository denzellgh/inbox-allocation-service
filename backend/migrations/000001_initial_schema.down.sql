-- Drop tables in reverse order of creation (respect foreign keys)

DROP TABLE IF EXISTS grace_period_assignments;
DROP TABLE IF EXISTS conversation_labels;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS conversation_refs;
DROP TABLE IF EXISTS operator_status;
DROP TABLE IF EXISTS operator_inbox_subscriptions;
DROP TABLE IF EXISTS operators;
DROP TABLE IF EXISTS inboxes;
DROP TABLE IF EXISTS tenants;

-- Drop enum types
DROP TYPE IF EXISTS grace_period_reason;
DROP TYPE IF EXISTS conversation_state;
DROP TYPE IF EXISTS operator_status_type;
DROP TYPE IF EXISTS operator_role;
