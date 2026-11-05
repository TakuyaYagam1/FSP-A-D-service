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

-- name: ListCasesByOwner :many
SELECT id, owner_id, kind, origin, destination, created_at
FROM cases
WHERE owner_id = @owner_id
ORDER BY created_at DESC, id DESC
LIMIT 21 OFFSET @page_offset;

