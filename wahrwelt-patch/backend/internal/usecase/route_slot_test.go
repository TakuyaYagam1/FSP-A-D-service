package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

type routeSlotCasesStub struct{}

func (routeSlotCasesStub) CreateCase(_ context.Context, item domain.Case) (domain.Case, error) {
	return item, nil
}

func (routeSlotCasesStub) GetCase(context.Context, uuid.UUID) (domain.Case, error) {
	return domain.Case{}, domain.ErrNotFound
}

func (routeSlotCasesStub) ListCases(context.Context, uuid.UUID, int32) ([]domain.Case, error) {
	return nil, nil
}

type routeSlotUsersStub struct{}

func (routeSlotUsersStub) CreateUser(_ context.Context, item domain.User) (domain.User, error) {
	return item, nil
}

func (routeSlotUsersStub) GetUserByID(context.Context, uuid.UUID) (domain.User, error) {
	return domain.User{}, domain.ErrNotFound
}

func (routeSlotUsersStub) GetUserByUsername(context.Context, string) (domain.User, error) {
	return domain.User{}, domain.ErrNotFound
}

func (routeSlotUsersStub) ListUsers(context.Context) ([]domain.User, error) {
	return nil, nil
}

type routeSlotRepositoryStub struct {
	item domain.RouteSlot
}

func (r *routeSlotRepositoryStub) CreateRouteSlot(_ context.Context, item domain.RouteSlot) (domain.RouteSlot, error) {
	r.item = item
	return item, nil
}

func (r *routeSlotRepositoryStub) GetRouteSlot(_ context.Context, _ uuid.UUID) (domain.RouteSlot, error) {
	return r.item, nil
}

func (r *routeSlotRepositoryStub) ListRouteSlots(context.Context, uuid.UUID, uuid.UUID) ([]domain.RouteSlot, error) {
	return []domain.RouteSlot{r.item}, nil
}

type countingApprovalStoreStub struct {
	state domain.ApprovalState
}

func (s *countingApprovalStoreStub) RecordApproval(_ context.Context, _ uuid.UUID, actorID uuid.UUID, role domain.ApprovalRole, _ time.Duration) (domain.ApprovalState, error) {
	switch role {
	case domain.RoleCarrier:
		if s.state.CarrierID != (uuid.UUID{}) && s.state.CarrierID != actorID {
			return domain.ApprovalState{}, domain.ErrForbidden
		}
		s.state.CarrierID = actorID
	case domain.RoleInspector:
		if s.state.InspectorID != (uuid.UUID{}) && s.state.InspectorID != actorID {
			return domain.ApprovalState{}, domain.ErrForbidden
		}
		s.state.InspectorID = actorID
	default:
		return domain.ApprovalState{}, domain.ErrInvalidInput
	}
	return s.state, nil
}

func (s *countingApprovalStoreStub) GetApprovals(context.Context, uuid.UUID) (domain.ApprovalState, error) {
	if s.state.CarrierID == (uuid.UUID{}) && s.state.InspectorID == (uuid.UUID{}) {
		return domain.ApprovalState{}, domain.ErrNotFound
	}
	return s.state, nil
}

func TestApproveSameRoleDoesNotCompleteSlot(t *testing.T) {
	carrierID := uuid.New()
	inspectorID := uuid.New()
	slot := domain.RouteSlot{ID: uuid.New(), CarrierID: carrierID, InspectorID: inspectorID}
	approvals := &countingApprovalStoreStub{}
	service := NewRouteSlot(routeSlotCasesStub{}, routeSlotUsersStub{}, &routeSlotRepositoryStub{item: slot}, approvals, time.Minute)

	if _, err := service.Approve(context.Background(), carrierID, slot.ID, domain.RoleCarrier); err != nil {
		t.Fatalf("first carrier approval error = %v", err)
	}
	result, err := service.Approve(context.Background(), carrierID, slot.ID, domain.RoleCarrier)
	if err != nil {
		t.Fatalf("repeated carrier approval error = %v", err)
	}
	if result.Ready || result.InspectorApproved {
		t.Fatalf("repeated carrier approval completed slot: %+v", result)
	}
}

func TestApproveDifferentRolesCompletesSlot(t *testing.T) {
	carrierID := uuid.New()
	inspectorID := uuid.New()
	slot := domain.RouteSlot{ID: uuid.New(), CarrierID: carrierID, InspectorID: inspectorID}
	approvals := &countingApprovalStoreStub{}
	service := NewRouteSlot(routeSlotCasesStub{}, routeSlotUsersStub{}, &routeSlotRepositoryStub{item: slot}, approvals, time.Minute)

	if _, err := service.Approve(context.Background(), carrierID, slot.ID, domain.RoleCarrier); err != nil {
		t.Fatalf("carrier approval error = %v", err)
	}
	result, err := service.Approve(context.Background(), inspectorID, slot.ID, domain.RoleInspector)
	if err != nil {
		t.Fatalf("inspector approval error = %v", err)
	}
	if !result.Ready || !result.CarrierApproved || !result.InspectorApproved {
		t.Fatalf("approvals did not complete slot: %+v", result)
	}
}
