-- name: InsertUser :one
INSERT INTO users (name, email, password)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByTelegramChatID :one
SELECT * FROM users
WHERE telegram_chat_id = $1;

-- name: UpdateUserTelegramChatID :exec
UPDATE users
SET telegram_chat_id = $2, updated_at = NOW()
WHERE id = $1;

-- name: GetUserWithRole :one
SELECT
    u.id,
    u.name,
    u.email,
    u.password,
    u.telegram_chat_id,
    u.role_id,
    r.name AS role_name
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
WHERE u.email = $1;

-- name: GetUserPermissions :many
SELECT p.name
FROM users u
JOIN roles r ON u.role_id = r.id
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE u.email = $1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;
