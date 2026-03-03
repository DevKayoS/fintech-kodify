-- name: InsertExpense :one
INSERT INTO expenses (user_id, category_id, amount, description, occurred_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetExpenseByID :one
SELECT
    e.id,
    e.user_id,
    e.amount,
    e.description,
    e.occurred_at,
    e.created_at,
    ec.slug AS category_slug,
    ec.name AS category_name
FROM expenses e
LEFT JOIN expense_categories ec ON e.category_id = ec.id
WHERE e.id = $1 AND e.user_id = $2;

-- name: ListExpensesByUser :many
SELECT
    e.id,
    e.user_id,
    e.amount,
    e.description,
    e.occurred_at,
    e.created_at,
    ec.slug AS category_slug,
    ec.name AS category_name
FROM expenses e
LEFT JOIN expense_categories ec ON e.category_id = ec.id
WHERE e.user_id = $1
ORDER BY e.occurred_at DESC;

-- name: ListExpensesByUserAndPeriod :many
SELECT
    e.id,
    e.user_id,
    e.amount,
    e.description,
    e.occurred_at,
    e.created_at,
    ec.slug AS category_slug,
    ec.name AS category_name
FROM expenses e
LEFT JOIN expense_categories ec ON e.category_id = ec.id
WHERE e.user_id = $1
  AND e.occurred_at >= $2
  AND e.occurred_at < $3
ORDER BY e.occurred_at DESC;

-- name: DeleteExpense :exec
DELETE FROM expenses WHERE id = $1 AND user_id = $2;

-- name: GetExpenseCategoryBySlug :one
SELECT * FROM expense_categories WHERE slug = $1;

-- name: ListExpenseCategories :many
SELECT * FROM expense_categories ORDER BY name;
