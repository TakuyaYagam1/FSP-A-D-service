-- name: CreateUser :one
INSERT INTO users (id, username, password_hash)
    VALUES (@id, @username, @password_hash)
RETURNING
    id, username, password_hash, created_at;

-- name: GetUserByID :one
SELECT
    id,
    username,
    password_hash,
    created_at
FROM
    users
WHERE
    id = @id;

-- name: GetUserByUsername :one
SELECT
    id,
    username,
    password_hash,
    created_at
FROM
    users
WHERE
    username = @username;

-- name: ListUsers :many
SELECT
    id,
    username,
    created_at
FROM
    users
ORDER BY
    username
LIMIT 500;
