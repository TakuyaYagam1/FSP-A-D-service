package domain

import (
	"time"

	"uuid"
)

type Waybill struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Number      string
	ArchiveCode string
	State       string
	CreatedAt   time.Time
}

type WaybillCard struct {
	Number string
	State  string
}
