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
	Count int64
}

func (state ApprovalState) Complete() bool {
	return state.Count >= 2
}

type ApprovalResult struct {
	CarrierApproved   bool
	InspectorApproved bool
	Ready             bool
}

func NewApprovalResult(state ApprovalState, role ApprovalRole) ApprovalResult {
	ready := state.Complete()
	return ApprovalResult{
		CarrierApproved:   ready || role == RoleCarrier,
		InspectorApproved: ready || role == RoleInspector,
		Ready:             ready,
	}
}
