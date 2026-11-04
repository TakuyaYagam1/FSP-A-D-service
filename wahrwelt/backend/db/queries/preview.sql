-- name: CreatePreview :one
INSERT INTO previews (id, owner_id)
    VALUES (@id, @owner_id)
RETURNING
    id, owner_id, created_at;

-- name: GetPreviewForOwner :one
SELECT
    id,
    owner_id,
    created_at
FROM
    previews
WHERE
    id = @id
    AND owner_id = @owner_id;

