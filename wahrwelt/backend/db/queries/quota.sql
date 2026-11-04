-- name: CreateQuota :one
INSERT INTO quotas (id, owner_id, quota_number, quota_code, volume, expires_at)
    VALUES (@id, @owner_id, @quota_number, @quota_code, @volume, @expires_at)
RETURNING
    id, owner_id, quota_number, quota_code, volume, expires_at, created_at;

-- name: GetQuotaByNumber :one
SELECT
    id,
    owner_id,
    quota_number,
    quota_code,
    volume,
    expires_at,
    created_at
FROM
    quotas
WHERE
    quota_number = @quota_number;

