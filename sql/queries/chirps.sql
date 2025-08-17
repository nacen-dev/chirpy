-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetChirps :many
SELECT id, created_at, updated_at, body, user_id
FROM chirps
WHERE user_id = @author_id OR @author_id = '00000000-0000-0000-0000-000000000000'
ORDER BY
    CASE 
        WHEN @order_by::text IS NULL 
          OR @order_by::text = 'asc' 
        THEN created_at 
    END ASC,
    CASE 
        WHEN @order_by::text = 'desc' 
        THEN created_at 
    END DESC;

-- name: GetChirpById :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirpById :exec
DELETE FROM chirps
WHERE id = $1;