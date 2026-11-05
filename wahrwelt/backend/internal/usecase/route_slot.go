package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"uuid"
)

type RouteSlot struct {
	cases     outbound.CaseRepository
	users     outbound.UserRepository
	slots     outbound.RouteSlotRepository
	approvals outbound.ApprovalStore
	ttl       time.Duration
}

func NewRouteSlot(cases outbound.CaseRepository, users outbound.UserRepository, slots outbound.RouteSlotRepository, approvals outbound.ApprovalStore, ttl time.Duration) *RouteSlot {
	return &RouteSlot{cases: cases, users: users, slots: slots, approvals: approvals, ttl: ttl}
}

func (s *RouteSlot) Create(ctx context.Context, ownerID, caseID uuid.UUID, carrierName, inspectorName, code string) (domain.RouteSlot, error) {
	if ownerID == (uuid.UUID{}) || caseID == (uuid.UUID{}) || s.ttl <= 0 {
		return domain.RouteSlot{}, domain.ErrInvalidInput
	}
	code, err := domain.ValidateRouteText(code)
	if err != nil {
		return domain.RouteSlot{}, err
	}
	carrierName, err = domain.NormalizeUsername(carrierName)
	if err != nil {
		return domain.RouteSlot{}, err
	}
	inspectorName, err = domain.NormalizeUsername(inspectorName)
	if err != nil {
		return domain.RouteSlot{}, err
	}
	item, err := s.cases.GetCase(ctx, caseID)
	if err != nil {
		return domain.RouteSlot{}, fmt.Errorf("find case: %w", err)
	}
	if item.OwnerID != ownerID {
		return domain.RouteSlot{}, domain.ErrNotFound
	}
	carrier, err := s.users.GetUserByUsername(ctx, carrierName)
	if err != nil {
		return domain.RouteSlot{}, participantError(err)
	}
	inspector, err := s.users.GetUserByUsername(ctx, inspectorName)
	if err != nil {
		return domain.RouteSlot{}, participantError(err)
	}
	if carrier.ID == inspector.ID || carrier.ID == (uuid.UUID{}) || inspector.ID == (uuid.UUID{}) {
		return domain.RouteSlot{}, domain.ErrInvalidInput
	}
	slot := domain.RouteSlot{
		ID:          uuid.New(),
		CaseID:      caseID,
		OwnerID:     ownerID,
		CarrierID:   carrier.ID,
		InspectorID: inspector.ID,
		Code:        code,
		CreatedAt:   time.Now().UTC(),
	}
	created, err := s.slots.CreateRouteSlot(ctx, slot)
	if err != nil {
		return domain.RouteSlot{}, fmt.Errorf("create route slot: %w", err)
	}
	created.CarrierUsername = carrier.Username
	created.InspectorUsername = inspector.Username
	return created, nil
}

func (s *RouteSlot) Approve(ctx context.Context, actorID, slotID uuid.UUID, role domain.ApprovalRole) (domain.ApprovalResult, error) {
	if actorID == (uuid.UUID{}) || slotID == (uuid.UUID{}) || !role.Valid() {
		return domain.ApprovalResult{}, domain.ErrInvalidInput
	}
	slot, err := s.slots.GetRouteSlot(ctx, slotID)
	if err != nil {
		return domain.ApprovalResult{}, fmt.Errorf("find route slot: %w", err)
	}
	if (role == domain.RoleCarrier && slot.CarrierID != actorID) || (role == domain.RoleInspector && slot.InspectorID != actorID) {
		return domain.ApprovalResult{}, domain.ErrForbidden
	}
	state, err := s.approvals.Increment(ctx, slotID, s.ttl)
	if err != nil {
		return domain.ApprovalResult{}, fmt.Errorf("record route slot approval: %w", err)
	}
	return domain.NewApprovalResult(state, role), nil
}

func (s *RouteSlot) Status(ctx context.Context, actorID, slotID uuid.UUID) (domain.RouteSlotStatus, error) {
	if actorID == (uuid.UUID{}) || slotID == (uuid.UUID{}) {
		return domain.RouteSlotStatus{}, domain.ErrNotFound
	}
	slot, err := s.slots.GetRouteSlot(ctx, slotID)
	if err != nil {
		return domain.RouteSlotStatus{}, fmt.Errorf("find route slot: %w", err)
	}
	if actorID != slot.OwnerID && actorID != slot.CarrierID && actorID != slot.InspectorID {
		return domain.RouteSlotStatus{}, domain.ErrNotFound
	}
	return s.currentStatus(ctx, slot)
}

func (s *RouteSlot) currentStatus(ctx context.Context, slot domain.RouteSlot) (domain.RouteSlotStatus, error) {
	state, err := s.approvals.GetApprovals(ctx, slot.ID)
	if errors.Is(err, domain.ErrNotFound) {
		state = domain.ApprovalState{}
	} else if err != nil {
		return domain.RouteSlotStatus{}, fmt.Errorf("read route slot approvals: %w", err)
	}
	slot.Code = ""
	return domain.RouteSlotStatus{Slot: slot, ApprovalCount: state.Count, Ready: state.Complete()}, nil
}

func (s *RouteSlot) List(ctx context.Context, actorID, caseID uuid.UUID) ([]domain.RouteSlotStatus, error) {
	if actorID == (uuid.UUID{}) || caseID == (uuid.UUID{}) {
		return nil, domain.ErrNotFound
	}
	item, err := s.cases.GetCase(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("find case: %w", err)
	}
	if item.OwnerID != actorID {
		return nil, domain.ErrNotFound
	}
	slots, err := s.slots.ListRouteSlots(ctx, caseID, actorID)
	if err != nil {
		return nil, fmt.Errorf("list route slots: %w", err)
	}
	result := make([]domain.RouteSlotStatus, 0, len(slots))
	for _, slot := range slots {
		if slot.OwnerID != actorID || slot.CaseID != caseID {
			return nil, domain.ErrNotFound
		}
		status, err := s.currentStatus(ctx, slot)
		if err != nil {
			return nil, err
		}
		result = append(result, status)
	}
	return result, nil
}

func (s *RouteSlot) Code(ctx context.Context, actorID, slotID uuid.UUID) (string, error) {
	if actorID == (uuid.UUID{}) || slotID == (uuid.UUID{}) {
		return "", domain.ErrNotFound
	}
	slot, err := s.slots.GetRouteSlot(ctx, slotID)
	if err != nil {
		return "", fmt.Errorf("find route slot: %w", err)
	}
	if slot.OwnerID == actorID {
		return slot.Code, nil
	}
	if actorID != slot.CarrierID && actorID != slot.InspectorID {
		return "", domain.ErrNotFound
	}
	state, err := s.approvals.GetApprovals(ctx, slotID)
	if errors.Is(err, domain.ErrNotFound) {
		return "", domain.ErrApprovalNotReady
	}
	if err != nil {
		return "", fmt.Errorf("load route slot approvals: %w", err)
	}
	if !state.Complete() {
		return "", domain.ErrApprovalNotReady
	}
	return slot.Code, nil
}

func participantError(err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrInvalidInput
	}
	return fmt.Errorf("find route slot participant: %w", err)
}
