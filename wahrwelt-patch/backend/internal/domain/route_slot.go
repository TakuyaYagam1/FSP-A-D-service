package domain

import (
	"time"

	"uuid"
)

type ApprovalRole string

const (
	RoleCarrier   ApprovalRole = "carrier"
	RoleInspector ApprovalRole = "inspector"
)

func (role ApprovalRole) Valid() bool {
	return role == RoleCarrier || role == RoleInspector
}

type RouteSlot struct {
	ID                uuid.UUID
	CaseID            uuid.UUID
	OwnerID           uuid.UUID
	CarrierID         uuid.UUID
	InspectorID       uuid.UUID
	CarrierUsername   string
	InspectorUsername string
	Code              string
	CreatedAt         time.Time
}

type ApprovalState struct {
	CarrierID   uuid.UUID
	InspectorID uuid.UUID
}

func (state ApprovalState) Complete() bool {
	return state.CarrierID != (uuid.UUID{}) && state.InspectorID != (uuid.UUID{})
}

type ApprovalResult struct {
	CarrierApproved   bool
	InspectorApproved bool
	Ready             bool
}

func NewApprovalResult(state ApprovalState) ApprovalResult {
	ready := state.Complete()
	return ApprovalResult{
		CarrierApproved:   state.CarrierID != (uuid.UUID{}),
		InspectorApproved: state.InspectorID != (uuid.UUID{}),
		Ready:             ready,
	}
}
