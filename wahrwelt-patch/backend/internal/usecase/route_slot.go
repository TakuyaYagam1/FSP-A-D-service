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
	// Каждая подпись привязана к назначенному пользователю и его роли
	state, err := s.approvals.RecordApproval(ctx, slotID, actorID, role, s.ttl)
	if err != nil {
		return domain.ApprovalResult{}, fmt.Errorf("record route slot approval: %w", err)
	}
	if (state.CarrierID != (uuid.UUID{}) && state.CarrierID != slot.CarrierID) || (state.InspectorID != (uuid.UUID{}) && state.InspectorID != slot.InspectorID) {
		return domain.ApprovalResult{}, domain.ErrForbidden
	}
	return domain.NewApprovalResult(state), nil
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
	// Код выдается только при двух подписях назначенных участников
	if !state.Complete() || state.CarrierID != slot.CarrierID || state.InspectorID != slot.InspectorID {
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
