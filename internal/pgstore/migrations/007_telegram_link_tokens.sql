-- Write your migrate up statements here

CREATE TABLE telegram_link_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(36) UNIQUE NOT NULL,  -- UUID v4
    expires_at TIMESTAMPTZ NOT NULL,          -- created_at + 5 min
    used_at    TIMESTAMPTZ,                   -- NULL = não usado
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_link_tokens_token ON telegram_link_tokens(token);

---- create above / drop below ----

DROP TABLE IF EXISTS telegram_link_tokens;
