CREATE TABLE api_keys (
    id                BIGSERIAL PRIMARY KEY,
    key_hash          CHAR(64) NOT NULL,
    key_prefix        VARCHAR(12) NOT NULL,
    holder_email      VARCHAR(255) NOT NULL,
    label             VARCHAR(255),
    access_request_id BIGINT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at      TIMESTAMPTZ,
    revoked_at        TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_api_keys_key_hash ON api_keys (key_hash);
