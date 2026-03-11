-- Write your migrate up statements here

CREATE TABLE investments (
    id                 BIGSERIAL PRIMARY KEY,
    user_id            BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    investment_type_id INT NOT NULL REFERENCES investment_types(id),
    amount             BIGINT NOT NULL, -- em centavos
    description        VARCHAR(500),
    invested_at        TIMESTAMPTZ DEFAULT NOW(),
    created_at         TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_investments_user_id    ON investments(user_id);
CREATE INDEX idx_investments_invested_at ON investments(invested_at);

---- create above / drop below ----

DROP TABLE IF EXISTS investments;
