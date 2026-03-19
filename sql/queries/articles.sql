-- name: CreateArticle :one
INSERT INTO articles (title, content, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetArticles :many
SELECT * FROM articles
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetArticleByID :one
SELECT * FROM articles
WHERE id = $1;

-- name: UpdateArticle :one
UPDATE articles SET title = $2, content = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteArticle :exec
DELETE FROM articles
WHERE id = $1;