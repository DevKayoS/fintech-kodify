CREATE TABLE telegram_conversations (
    chat_id    BIGINT PRIMARY KEY,
    step       VARCHAR(50)  NOT NULL,
    data       TEXT         NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

---- create above / drop below ----

DROP TABLE IF EXISTS telegram_conversations;
