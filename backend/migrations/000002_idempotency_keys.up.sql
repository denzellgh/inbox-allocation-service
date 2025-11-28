-- Idempotency keys table for deduplication
CREATE TABLE IF NOT EXISTS idempotency_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    request_hash VARCHAR(64),
    response_status INT NOT NULL,
    response_body JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    
    CONSTRAINT uq_idempotency_tenant_key UNIQUE(tenant_id, key)
);

-- Index for lookup by tenant and key
CREATE INDEX idx_idempotency_keys_tenant_key ON idempotency_keys(tenant_id, key);

-- Index for cleanup of expired records
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);

COMMENT ON TABLE idempotency_keys IS 'Stores idempotency keys with cached responses for deduplication';
COMMENT ON COLUMN idempotency_keys.key IS 'Client-provided idempotency key (UUID recommended)';
COMMENT ON COLUMN idempotency_keys.request_hash IS 'SHA256 hash of request body for validation (optional)';
COMMENT ON COLUMN idempotency_keys.response_status IS 'HTTP status code of the original response';
COMMENT ON COLUMN idempotency_keys.response_body IS 'Full JSON response body';
COMMENT ON COLUMN idempotency_keys.expires_at IS 'When this record can be cleaned up';
