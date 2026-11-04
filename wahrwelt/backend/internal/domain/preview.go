package domain

import (
	"time"

	"uuid"
)

type Preview struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	CreatedAt time.Time
}

type PreviewResult struct {
	ID     uuid.UUID
	State  string
	PNG    []byte
	Stdout string
	Stderr string
}
