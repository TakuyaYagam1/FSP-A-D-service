package domain

import (
	"time"

	"uuid"
)

type Certificate struct {
	ID           uuid.UUID
	OwnerID      uuid.UUID
	Serial       string
	OriginPhrase string
	CreatedAt    time.Time
}

type CertificateSummary struct {
	ID           uuid.UUID
	OwnerID      uuid.UUID
	Serial       string
	OriginPhrase string
	CreatedAt    time.Time
}
