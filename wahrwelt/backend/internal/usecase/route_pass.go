package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"uuid"
)

type RoutePass struct {
	slots  *RouteSlot
	passes outbound.RoutePassRepository
}

func NewRoutePass(slots *RouteSlot, passes outbound.RoutePassRepository) *RoutePass {
	return &RoutePass{slots: slots, passes: passes}
}

func (s *RoutePass) Get(ctx context.Context, actorID, slotID uuid.UUID) (domain.RoutePassState, error) {
	status, err := s.slots.Status(ctx, actorID, slotID)
	if err != nil {
		return domain.RoutePassState{}, err
	}
	pass, err := s.passes.GetRoutePass(ctx, slotID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.RoutePassState{Ready: status.Ready}, nil
	}
	if err != nil {
		return domain.RoutePassState{}, fmt.Errorf("read route pass: %w", err)
	}
	return domain.RoutePassState{Ready: status.Ready, Pass: &pass}, nil
}

func (s *RoutePass) Issue(ctx context.Context, actorID, slotID uuid.UUID) (domain.RoutePassState, error) {
	status, err := s.slots.Status(ctx, actorID, slotID)
	if err != nil {
		return domain.RoutePassState{}, err
	}
	if status.Slot.OwnerID != actorID {
		return domain.RoutePassState{}, domain.ErrForbidden
	}
	pass, err := s.passes.GetRoutePass(ctx, slotID)
	if err == nil {
		return domain.RoutePassState{Ready: status.Ready, Pass: &pass}, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.RoutePassState{}, fmt.Errorf("read route pass: %w", err)
	}
	if !status.Ready {
		return domain.RoutePassState{}, domain.ErrApprovalNotReady
	}
	pass, err = s.passes.CreateRoutePass(ctx, slotID)
	if err != nil {
		return domain.RoutePassState{}, fmt.Errorf("issue route pass: %w", err)
	}
	return domain.RoutePassState{Ready: true, Pass: &pass}, nil
}
