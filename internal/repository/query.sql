-- name: SqlcCreateFilesItem :one
INSERT INTO files (user_id, parent_id, attachment_id, created_at, type, name, original_name)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;
