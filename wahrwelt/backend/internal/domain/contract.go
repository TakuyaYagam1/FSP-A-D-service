package domain

import (
	"time"

	"uuid"
)

type Contract struct {
	ID         uuid.UUID
	OwnerID    uuid.UUID
	Number     string
	CrewName   string
	Service    string
	CrewPhrase string
	StartsAt   time.Time
	EndsAt     time.Time
	CreatedAt  time.Time
}

type ContractCard struct {
	Number   string
	CrewName string
	Service  string
	StartsAt time.Time
	EndsAt   time.Time
}

type ContractPreview struct {
	Rendered string
}
