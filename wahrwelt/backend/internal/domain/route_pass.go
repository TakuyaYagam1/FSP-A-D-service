package domain

import (
	"time"

	"uuid"
)

type RoutePass struct {
	SlotID            uuid.UUID
	CaseID            uuid.UUID
	Kind              string
	Origin            string
	Destination       string
	CarrierUsername   string
	InspectorUsername string
	IssuedAt          time.Time
}

type RoutePassState struct {
	Ready bool
	Pass  *RoutePass
}
