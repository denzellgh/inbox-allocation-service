-- Enable UUID extension (backup, already done in init script)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- ENUM TYPES
-- ============================================================================

CREATE TYPE operator_role AS ENUM ('OPERATOR', 'MANAGER', 'ADMIN');
CREATE TYPE operator_status_type AS ENUM ('AVAILABLE', 'OFFLINE');
CREATE TYPE conversation_state AS ENUM ('QUEUED', 'ALLOCATED', 'RESOLVED');
CREATE TYPE grace_period_reason AS ENUM ('OFFLINE', 'MANUAL');

-- ============================================================================
-- TABLE: tenants
-- ============================================================================
-- Stores tenant configuration including priority weights for allocation scoring.
-- priority_weight_alpha: weight for normalized message count
-- priority_weight_beta: weight for normalized delay

CREATE TABLE tenants (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    priority_weight_alpha DECIMAL(5,4) NOT NULL DEFAULT 0.5,
    priority_weight_beta DECIMAL(5,4) NOT NULL DEFAULT 0.5,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID
);

-- Index for tenant lookup by name (unique per system)
CREATE UNIQUE INDEX idx_tenants_name ON tenants(name);

-- ============================================================================
-- TABLE: inboxes
-- ============================================================================
-- Each inbox represents a phone number channel for a tenant.
-- phone_number is unique per tenant.

CREATE TABLE inboxes (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    phone_number VARCHAR(20) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique constraint: one phone number per tenant
CREATE UNIQUE INDEX idx_inboxes_tenant_phone ON inboxes(tenant_id, phone_number);

-- Index for listing inboxes by tenant
CREATE INDEX idx_inboxes_tenant_id ON inboxes(tenant_id);

-- ============================================================================
-- TABLE: operators
-- ============================================================================
-- Operators are users who handle conversations. Role determines permissions.

CREATE TABLE operators (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role operator_role NOT NULL DEFAULT 'OPERATOR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for listing operators by tenant
CREATE INDEX idx_operators_tenant_id ON operators(tenant_id);

-- Index for filtering by role within tenant
CREATE INDEX idx_operators_tenant_role ON operators(tenant_id, role);

-- ============================================================================
-- TABLE: operator_inbox_subscriptions
-- ============================================================================
-- Junction table: which operators are subscribed to which inboxes.
-- An operator can only be allocated conversations from subscribed inboxes.

CREATE TABLE operator_inbox_subscriptions (
    id UUID PRIMARY KEY,
    operator_id UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    inbox_id UUID NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique constraint: one subscription per operator-inbox pair
CREATE UNIQUE INDEX idx_subscriptions_operator_inbox ON operator_inbox_subscriptions(operator_id, inbox_id);

-- Index for finding subscriptions by operator
CREATE INDEX idx_subscriptions_operator_id ON operator_inbox_subscriptions(operator_id);

-- Index for finding subscriptions by inbox
CREATE INDEX idx_subscriptions_inbox_id ON operator_inbox_subscriptions(inbox_id);

-- ============================================================================
-- TABLE: operator_status
-- ============================================================================
-- Tracks current status of each operator (AVAILABLE or OFFLINE).
-- One status record per operator.

CREATE TABLE operator_status (
    id UUID PRIMARY KEY,
    operator_id UUID NOT NULL UNIQUE REFERENCES operators(id) ON DELETE CASCADE,
    status operator_status_type NOT NULL DEFAULT 'OFFLINE',
    last_status_change_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for finding operators by status (useful for allocation)
CREATE INDEX idx_operator_status_status ON operator_status(status);

-- ============================================================================
-- TABLE: conversation_refs
-- ============================================================================
-- Core table: references to conversations with allocation metadata.
-- external_conversation_id links to the orchestrator system.
-- priority_score is pre-calculated for efficient sorting.

CREATE TABLE conversation_refs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    inbox_id UUID NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
    external_conversation_id VARCHAR(255) NOT NULL,
    customer_phone_number VARCHAR(20) NOT NULL,
    state conversation_state NOT NULL DEFAULT 'QUEUED',
    assigned_operator_id UUID REFERENCES operators(id) ON DELETE SET NULL,
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    message_count INTEGER NOT NULL DEFAULT 0,
    priority_score DECIMAL(10,6) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- CRITICAL INDEX: Used by allocation query (SELECT FOR UPDATE SKIP LOCKED)
-- Filters: tenant_id, inbox_id (from subscriptions), state = 'QUEUED'
-- Orders by: priority_score DESC, last_message_at ASC
CREATE INDEX idx_conversations_allocation ON conversation_refs(
    tenant_id,
    inbox_id,
    state,
    priority_score DESC,
    last_message_at ASC
) WHERE state = 'QUEUED';

-- Index for listing conversations by tenant and state
CREATE INDEX idx_conversations_tenant_state ON conversation_refs(tenant_id, state);

-- Index for finding conversations assigned to an operator
CREATE INDEX idx_conversations_operator ON conversation_refs(
    tenant_id,
    assigned_operator_id,
    state
) WHERE assigned_operator_id IS NOT NULL;

-- Index for phone number search (exact match)
CREATE INDEX idx_conversations_phone ON conversation_refs(tenant_id, customer_phone_number);

-- Unique constraint: external_conversation_id should be unique per tenant
CREATE UNIQUE INDEX idx_conversations_external_id ON conversation_refs(tenant_id, external_conversation_id);

-- ============================================================================
-- TABLE: labels
-- ============================================================================
-- Labels are scoped to inboxes, not global to tenant.

CREATE TABLE labels (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    inbox_id UUID NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7),
    created_by UUID REFERENCES operators(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique constraint: label name unique within inbox
CREATE UNIQUE INDEX idx_labels_inbox_name ON labels(inbox_id, name);

-- Index for listing labels by inbox
CREATE INDEX idx_labels_inbox_id ON labels(tenant_id, inbox_id);

-- ============================================================================
-- TABLE: conversation_labels
-- ============================================================================
-- Junction table: which labels are attached to which conversations.

CREATE TABLE conversation_labels (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversation_refs(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique constraint: one attachment per conversation-label pair
CREATE UNIQUE INDEX idx_conv_labels_conv_label ON conversation_labels(conversation_id, label_id);

-- Index for finding labels by conversation
CREATE INDEX idx_conv_labels_conversation ON conversation_labels(conversation_id);

-- Index for finding conversations by label
CREATE INDEX idx_conv_labels_label ON conversation_labels(label_id);

-- ============================================================================
-- TABLE: grace_period_assignments
-- ============================================================================
-- Temporary records created when an operator goes OFFLINE.
-- Conversations are not released immediately; they wait until expires_at.

CREATE TABLE grace_period_assignments (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversation_refs(id) ON DELETE CASCADE,
    operator_id UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    reason grace_period_reason NOT NULL DEFAULT 'OFFLINE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique constraint: one grace period per conversation
CREATE UNIQUE INDEX idx_grace_conversation ON grace_period_assignments(conversation_id);

-- CRITICAL INDEX: Used by grace period worker to find expired assignments
CREATE INDEX idx_grace_expires ON grace_period_assignments(expires_at)
WHERE expires_at IS NOT NULL;

-- Index for finding grace periods by operator (cleanup when returning online)
CREATE INDEX idx_grace_operator ON grace_period_assignments(operator_id);
