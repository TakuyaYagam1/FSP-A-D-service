package domain

import (
	"time"

	"uuid"
)

type Quota struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Number    string
	QuotaCode string
	Volume    int64
	ExpiresAt time.Time
	CreatedAt time.Time
}

type QuotaSummary struct {
	Quota  Quota
	Fields []string
}
