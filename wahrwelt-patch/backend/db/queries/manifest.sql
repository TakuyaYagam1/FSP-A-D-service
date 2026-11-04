-- name: CreateManifest :one
INSERT INTO manifests (id, owner_id, manifest_number, origin, destination, status)
    VALUES (@id, @owner_id, @manifest_number, @origin, @destination, @status)
RETURNING
    id, owner_id, manifest_number, origin, destination, status, created_at;

-- name: CreateManifestRevision :one
INSERT INTO manifest_revisions (id, manifest_id, owner_id, status, sealed_route)
    VALUES (@id, @manifest_id, @owner_id, @status, @sealed_route)
RETURNING
    id, manifest_id, owner_id, status, sealed_route, created_at;

-- name: GetManifestByNumber :one
SELECT
    id,
    owner_id,
    manifest_number,
    origin,
    destination,
    status,
    created_at
FROM
    manifests
WHERE
    manifest_number = @manifest_number;

-- name: ListManifestRevisions :many
SELECT
    id,
    status,
    created_at
FROM
    manifest_revisions
WHERE
    manifest_id = @manifest_id
ORDER BY
    created_at,
    id;

-- name: GetManifestRevision :one
SELECT
    r.id,
    r.manifest_id,
    r.owner_id,
    m.manifest_number,
    r.status,
    r.sealed_route,
    r.created_at
FROM
    manifest_revisions AS r
    JOIN manifests AS m ON m.id = r.manifest_id
WHERE
    r.manifest_id = @manifest_id
    AND r.id = @revision_id;
