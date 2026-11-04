package domain

import (
	"time"

	"uuid"
)

type GatePass struct {
	ID             uuid.UUID
	OwnerID        uuid.UUID
	GateOperatorID uuid.UUID
	Number         string
	State          string
	GateWord       string
	CreatedAt      time.Time
}

type GatePassCard struct {
	Number string
	State  string
}
