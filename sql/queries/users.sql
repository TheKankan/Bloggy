-- name: CreateUser :one
INSERT INTO users (username, hashed_password)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserFromUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: GetUsernameFromID :one
SELECT username FROM users
WHERE id = $1;