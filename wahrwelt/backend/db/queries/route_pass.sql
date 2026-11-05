-- name: CreateRoutePass :one
INSERT INTO route_passes (route_slot_id)
VALUES ($1)
ON CONFLICT (route_slot_id) DO UPDATE SET route_slot_id = EXCLUDED.route_slot_id
RETURNING issued_at;

-- name: GetRoutePass :one
SELECT p.route_slot_id, p.issued_at, s.case_id, c.kind, c.origin, c.destination,
       carrier.username AS carrier_username, inspector.username AS inspector_username
FROM route_passes AS p
JOIN route_slots AS s ON s.id = p.route_slot_id
JOIN cases AS c ON c.id = s.case_id
JOIN users AS carrier ON carrier.id = s.carrier_id
JOIN users AS inspector ON inspector.id = s.inspector_id
WHERE p.route_slot_id = $1;
