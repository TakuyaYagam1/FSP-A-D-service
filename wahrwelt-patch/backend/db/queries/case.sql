-- name: CreateCase :one
INSERT INTO cases (id, owner_id, kind, origin, destination)
    VALUES (@id, @owner_id, @kind, @origin, @destination)
RETURNING
    id, owner_id, kind, origin, destination, created_at;

-- name: GetCase :one
SELECT
    id,
    owner_id,
    kind,
    origin,
    destination,
    created_at
FROM
    cases
WHERE
    id = @id;
