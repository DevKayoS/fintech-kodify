-- Write your migrate up statements here

CREATE TABLE expenses (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INT REFERENCES expense_categories(id),
    amount      BIGINT NOT NULL,       -- em centavos
    description VARCHAR(500),
    occurred_at TIMESTAMPTZ DEFAULT NOW(),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_expenses_user_id     ON expenses(user_id);
CREATE INDEX idx_expenses_occurred_at ON expenses(occurred_at);

---- create above / drop below ----

DROP TABLE IF EXISTS expenses;
