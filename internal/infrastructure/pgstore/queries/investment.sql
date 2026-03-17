-- name: InsertInvestment :one
INSERT INTO investments (user_id, investment_type_id, amount, description, invested_at, movement_type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetInvestmentByID :one
SELECT
    i.id,
    i.user_id,
    i.amount,
    i.description,
    i.invested_at,
    i.created_at,
    i.movement_type,
    it.slug AS type_slug,
    it.name AS type_name
FROM investments i
JOIN investment_types it ON i.investment_type_id = it.id
WHERE i.id = $1 AND i.user_id = $2;

-- name: ListInvestmentsByUser :many
SELECT
    i.id,
    i.user_id,
    i.amount,
    i.description,
    i.invested_at,
    i.created_at,
    i.movement_type,
    it.slug AS type_slug,
    it.name AS type_name
FROM investments i
JOIN investment_types it ON i.investment_type_id = it.id
WHERE i.user_id = $1
ORDER BY i.invested_at DESC;

-- name: ListInvestmentsByUserAndPeriod :many
SELECT
    i.id,
    i.user_id,
    i.amount,
    i.description,
    i.invested_at,
    i.created_at,
    i.movement_type,
    it.slug AS type_slug,
    it.name AS type_name
FROM investments i
JOIN investment_types it ON i.investment_type_id = it.id
WHERE i.user_id = $1
  AND i.invested_at >= $2
  AND i.invested_at < $3
ORDER BY i.invested_at DESC;

-- name: DeleteInvestment :exec
DELETE FROM investments WHERE id = $1 AND user_id = $2;

-- name: GetInvestmentTypeBySlug :one
SELECT * FROM investment_types WHERE slug = $1;

-- name: ListInvestmentTypes :many
SELECT * FROM investment_types ORDER BY name;

-- name: GetTotalInvestmentsByUser :one
SELECT COALESCE(SUM(
    CASE WHEN movement_type = 'withdrawal' THEN -amount ELSE amount END
), 0)::BIGINT AS total
FROM investments
WHERE user_id = $1;

-- name: GetInvestmentSummaryByPeriod :many
SELECT
    it.slug AS type_slug,
    it.name AS type_name,
    SUM(CASE WHEN i.movement_type = 'withdrawal' THEN -i.amount ELSE i.amount END) AS total
FROM investments i
JOIN investment_types it ON i.investment_type_id = it.id
WHERE i.user_id = $1
  AND i.invested_at >= $2
  AND i.invested_at < $3
GROUP BY it.id, it.slug, it.name
ORDER BY total DESC;
