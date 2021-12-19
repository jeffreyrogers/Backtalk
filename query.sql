-- name: CreateComment :one
INSERT INTO comments (
  slug, author, content
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetComments :many
SELECT * FROM comments
WHERE slug = $1
ORDER BY created;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1;

-- name: GetUser :one
SELECT * FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (
  email, hash, salt
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: DeleteOldSessions :exec
DELETE FROM sessions
WHERE last_seen < now() - interval '1 month';

-- name: GetSession :one
SELECT * FROM sessions
WHERE session_id = $1;

-- name: CreateSession :exec
INSERT INTO sessions (
  session_id, uid  
) VALUES (
  $1, $2
);