-- name: GetDBStatus :one
SELECT
    split_part(current_setting('server_version'), ' ', 1) AS version,
    current_setting('max_connections')::int AS max_connections,
    (SELECT version FROM schema_version ORDER BY version DESC LIMIT 1) AS last_migration
FROM (SELECT 1) AS dummy;
