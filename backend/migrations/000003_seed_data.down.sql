-- ============================================================================
-- SEED DATA ROLLBACK
-- ============================================================================
-- Removes all seed data in reverse order of dependencies.
-- ============================================================================

-- Delete grace_period_assignments first (depends on conversations and operators)
DELETE FROM grace_period_assignments WHERE id LIKE '90000000-0000-0000-0000-%';

-- Delete conversation_labels (depends on conversations and labels)
DELETE FROM conversation_labels WHERE id LIKE '80000000-0000-0000-0000-%';

-- Delete conversation_refs (depends on tenants, inboxes, operators)
DELETE FROM conversation_refs WHERE id LIKE '70000000-0000-0000-0000-%';

-- Delete labels (depends on tenants, inboxes, operators)
DELETE FROM labels WHERE id LIKE '60000000-0000-0000-0000-%';

-- Delete operator_status (depends on operators)
DELETE FROM operator_status WHERE id LIKE '50000000-0000-0000-0000-%';

-- Delete operator_inbox_subscriptions (depends on operators and inboxes)
DELETE FROM operator_inbox_subscriptions WHERE id LIKE '40000000-0000-0000-0000-%';

-- Delete operators (depends on tenants)
DELETE FROM operators WHERE id LIKE '30000000-0000-0000-0000-%';

-- Delete inboxes (depends on tenants)
DELETE FROM inboxes WHERE id LIKE '20000000-0000-0000-0000-%';

-- Delete tenants (root)
DELETE FROM tenants WHERE id LIKE '10000000-0000-0000-0000-%';
