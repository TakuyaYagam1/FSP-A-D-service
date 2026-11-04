-- name: CreateLicense :one
INSERT INTO export_licenses (id, owner_id, license_number, status, release_pin)
    VALUES (@id, @owner_id, @license_number, @status, @release_pin)
RETURNING
    id, owner_id, license_number, status, release_pin, created_at;

-- name: GetLicenseByNumber :one
SELECT
    id,
    owner_id,
    license_number,
    status,
    release_pin,
    created_at
FROM
    export_licenses
WHERE
    license_number = @license_number;
