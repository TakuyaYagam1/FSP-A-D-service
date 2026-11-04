package domain

import (
	"time"

	"uuid"
)

type ManifestStatus string

const ManifestStatusReceived ManifestStatus = "received"

type Manifest struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Number      string
	Status      ManifestStatus
	Origin      string
	Destination string
	CreatedAt   time.Time
}

type ManifestRevision struct {
	ID             uuid.UUID
	ManifestID     uuid.UUID
	OwnerID        uuid.UUID
	ManifestNumber string
	Status         ManifestStatus
	SealedRoute    string
	CreatedAt      time.Time
}

type RevisionSummary struct {
	ID        uuid.UUID
	Status    ManifestStatus
	CreatedAt time.Time
}
