-- name: CreateWaybill :one
INSERT INTO archive_waybills (id, owner_id, waybill_number, archive_code)
    VALUES (@id, @owner_id, @waybill_number, @archive_code)
RETURNING
    id, owner_id, waybill_number, archive_code, state, created_at;

-- name: GetWaybillByNumber :one
SELECT
    id,
    owner_id,
    waybill_number,
    archive_code,
    state,
    created_at
FROM
    archive_waybills
WHERE
    waybill_number = @waybill_number;
