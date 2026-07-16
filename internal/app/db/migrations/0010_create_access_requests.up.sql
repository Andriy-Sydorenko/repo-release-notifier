CREATE TABLE access_requests (
    id          BIGSERIAL PRIMARY KEY,
    email       VARCHAR(255) NOT NULL,
    reason      TEXT,
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_access_requests_pending_email ON access_requests (email) WHERE status = 'pending';
