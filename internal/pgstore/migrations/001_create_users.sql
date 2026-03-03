-- Write your migrate up statements here

CREATE TABLE users (
    id               BIGSERIAL PRIMARY KEY,
    name             VARCHAR(255) NOT NULL,
    email            VARCHAR(255) UNIQUE NOT NULL,
    password         VARCHAR(255) NOT NULL,
    telegram_chat_id BIGINT UNIQUE,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

---- create above / drop below ----

DROP TABLE IF EXISTS users;
