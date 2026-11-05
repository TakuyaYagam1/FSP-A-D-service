-- +goose Up
CREATE TABLE route_passes (
    route_slot_id uuid PRIMARY KEY REFERENCES route_slots (id) ON DELETE CASCADE,
    issued_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE route_passes;
