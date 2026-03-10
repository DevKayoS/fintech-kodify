-- name: InsertLinkToken :one
INSERT INTO telegram_link_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetValidLinkToken :one
SELECT * FROM telegram_link_tokens
WHERE token = $1
  AND expires_at > NOW()
  AND used_at IS NULL;

-- name: MarkLinkTokenUsed :exec
UPDATE telegram_link_tokens
SET used_at = NOW()
WHERE id = $1;
