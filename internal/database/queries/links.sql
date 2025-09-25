-- name: CreateLink :one
INSERT INTO links (user_id, file_path, expires_at, max_downloads)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLinkByID :one
SELECT * FROM links
WHERE id = $1 LIMIT 1;

-- name: ListLinksByUser :many
SELECT * FROM links
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteLink :exec
DELETE FROM links
WHERE id = $1;

-- name: IncrementDownload :one
UPDATE links
SET downloads = downloads + 1
WHERE id = $1
RETURNING *;

-- name: ExpiredLinks :many
SELECT * FROM links
WHERE expires_at < NOW();

-- name: ActiveLinksByUser :many
SELECT * FROM links
WHERE user_id = $1 AND expires_at >= NOW()
ORDER BY created_at DESC;
