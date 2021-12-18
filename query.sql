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
