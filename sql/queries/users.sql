-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET email = sqlc.arg(new_email)::text, hashed_password = sqlc.arg(new_password)::text, updated_at = NOW()
WHERE email = sqlc.arg(old_email)::text
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1;

-- name: UpgradeUserToChirpyRed :one
UPDATE users
SET is_chirpy_red = true, updated_at = NOW()
WHERE id = $1
RETURNING id, created_at, updated_at, email, is_chirpy_red;

