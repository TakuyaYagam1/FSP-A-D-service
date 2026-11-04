package domain

import (
	"time"

	"uuid"
)

type LicenseStatus string

const LicenseStatusActive LicenseStatus = "active"

type License struct {
	ID         uuid.UUID
	OwnerID    uuid.UUID
	Number     string
	Status     LicenseStatus
	ReleasePin string
	CreatedAt  time.Time
}

type LicenseCard struct {
	Number string
	Status LicenseStatus
}

type LicenseToken struct {
	Value     string
	ExpiresAt time.Time
}
