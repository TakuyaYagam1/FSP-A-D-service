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

// Сводка содержит только открытые значения и не несет quota_code
type QuotaSummary struct {
	Number    string
	Volume    int64
	ExpiresAt time.Time
	Fields    []string
}
