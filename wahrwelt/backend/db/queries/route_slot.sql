-- name: CreateRouteSlot :one
INSERT INTO route_slots (id, case_id, owner_id, carrier_id, inspector_id, slot_code)
    VALUES (@id, @case_id, @owner_id, @carrier_id, @inspector_id, @slot_code)
RETURNING
    id, case_id, owner_id, carrier_id, inspector_id, slot_code, created_at;

-- name: GetRouteSlot :one
SELECT
    id,
    case_id,
    owner_id,
    carrier_id,
    inspector_id,
    slot_code,
    created_at
FROM
    route_slots
WHERE
    id = @id;

