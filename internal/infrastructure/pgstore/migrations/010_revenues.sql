CREATE TABLE revenues (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount      BIGINT       NOT NULL,
    description TEXT,
    received_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_revenues_user_id ON revenues(user_id);
CREATE INDEX idx_revenues_received_at ON revenues(received_at);

---- create above / drop below ----

DROP TABLE IF EXISTS revenues;
