-- name: CreateCertificate :one
INSERT INTO certificates (id, owner_id, serial, origin_phrase)
    VALUES (@id, @owner_id, @serial, @origin_phrase)
RETURNING
    id, owner_id, serial, origin_phrase, created_at;

-- name: GetCertificateBySerial :one
SELECT
    id,
    owner_id,
    serial,
    origin_phrase,
    created_at
FROM
    certificates
WHERE
    serial = @serial;
