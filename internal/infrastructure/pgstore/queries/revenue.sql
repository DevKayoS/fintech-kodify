-- name: InsertRevenue :one
INSERT INTO revenues (user_id, amount, description, received_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, amount, description, received_at, created_at;

-- name: ListRevenuesByUser :many
SELECT id, user_id, amount, description, received_at, created_at
FROM revenues
WHERE user_id = $1
ORDER BY received_at DESC;

-- name: GetRevenueSummaryByPeriod :one
SELECT COALESCE(SUM(amount), 0)::BIGINT AS total
FROM revenues
WHERE user_id = $1
  AND received_at >= $2
  AND received_at < $3;
