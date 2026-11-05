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

-- name: ListRouteSlotsByCase :many
SELECT s.id, s.case_id, s.owner_id, s.carrier_id, s.inspector_id,
       c.username AS carrier_username, i.username AS inspector_username, s.created_at
FROM route_slots AS s
JOIN users AS c ON c.id = s.carrier_id
JOIN users AS i ON i.id = s.inspector_id
WHERE s.case_id = @case_id AND s.owner_id = @owner_id
ORDER BY s.created_at DESC, s.id DESC;

