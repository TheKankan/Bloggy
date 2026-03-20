-- name: CreateComment :one
INSERT INTO comments (content, user_id, article_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCommentsByArticleID :many
SELECT * FROM comments
WHERE article_id = $1
ORDER BY created_at ASC;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1;

-- name: GetCommentByID :one
SELECT * FROM comments
WHERE id = $1;