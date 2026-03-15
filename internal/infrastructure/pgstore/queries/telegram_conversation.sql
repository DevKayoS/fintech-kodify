-- name: GetConversationState :one
SELECT chat_id, step, data, updated_at
FROM telegram_conversations
WHERE chat_id = $1;

-- name: UpsertConversationState :exec
INSERT INTO telegram_conversations (chat_id, step, data, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (chat_id) DO UPDATE
SET step = EXCLUDED.step,
    data = EXCLUDED.data,
    updated_at = NOW();

-- name: DeleteConversationState :exec
DELETE FROM telegram_conversations WHERE chat_id = $1;
