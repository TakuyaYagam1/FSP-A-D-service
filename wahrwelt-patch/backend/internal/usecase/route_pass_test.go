package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

type passSlotStore struct{ slot domain.RouteSlot }

func (s *passSlotStore) CreateRouteSlot(_ context.Context, slot domain.RouteSlot) (domain.RouteSlot, error) {
	return slot, nil
}
func (s *passSlotStore) GetRouteSlot(_ context.Context, id uuid.UUID) (domain.RouteSlot, error) {
	if id != s.slot.ID {
		return domain.RouteSlot{}, domain.ErrNotFound
	}
	return s.slot, nil
}
func (s *passSlotStore) ListRouteSlots(context.Context, uuid.UUID, uuid.UUID) ([]domain.RouteSlot, error) {
	return []domain.RouteSlot{s.slot}, nil
}

type passApprovalStore struct{ state domain.ApprovalState }

func (s *passApprovalStore) RecordApproval(_ context.Context, _ uuid.UUID, actorID uuid.UUID, role domain.ApprovalRole, _ time.Duration) (domain.ApprovalState, error) {
	if role == domain.RoleCarrier {
		s.state.CarrierID = actorID
	} else {
		s.state.InspectorID = actorID
	}
	return s.state, nil
}
func (s *passApprovalStore) GetApprovals(context.Context, uuid.UUID) (domain.ApprovalState, error) {
	if s.state.CarrierID == (uuid.UUID{}) && s.state.InspectorID == (uuid.UUID{}) {
		return domain.ApprovalState{}, domain.ErrNotFound
	}
	return s.state, nil
}

type passRepository struct {
	pass   *domain.RoutePass
	issues int
}

func (r *passRepository) GetRoutePass(_ context.Context, _ uuid.UUID) (domain.RoutePass, error) {
	if r.pass == nil {
		return domain.RoutePass{}, domain.ErrNotFound
	}
	return *r.pass, nil
}
func (r *passRepository) CreateRoutePass(_ context.Context, slotID uuid.UUID) (domain.RoutePass, error) {
	r.issues++
	issued := domain.RoutePass{SlotID: slotID, CaseID: uuid.New(), Kind: "import", Origin: "Орел", Destination: "Тула", IssuedAt: time.Date(2026, time.November, 5, 17, 0, 0, 0, time.UTC)}
	r.pass = &issued
	return issued, nil
}

func TestRoutePassRejectsMismatchedRoleIdentity(t *testing.T) {
	ownerID, carrierID, inspectorID := uuid.New(), uuid.New(), uuid.New()
	slot := domain.RouteSlot{ID: uuid.New(), OwnerID: ownerID, CarrierID: carrierID, InspectorID: inspectorID}
	approvals := &passApprovalStore{state: domain.ApprovalState{CarrierID: carrierID, InspectorID: carrierID}}
	passes := &passRepository{}
	service := NewRoutePass(NewRouteSlot(nil, nil, &passSlotStore{slot: slot}, approvals, time.Minute), passes)
	if _, err := service.Issue(context.Background(), ownerID, slot.ID); !errors.Is(err, domain.ErrApprovalNotReady) {
		t.Fatalf("wrong inspector identity issued pass: %v", err)
	}
	if passes.issues != 0 {
		t.Fatalf("wrong identity stored a pass: %d", passes.issues)
	}
}

func TestRoutePassIssueRequiresReadyOwnerAndKeepsIssuedCopy(t *testing.T) {
	ownerID, carrierID, inspectorID, strangerID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	slot := domain.RouteSlot{ID: uuid.New(), CaseID: uuid.New(), OwnerID: ownerID, CarrierID: carrierID, InspectorID: inspectorID, Code: "private-code"}
	approvals := &passApprovalStore{state: domain.ApprovalState{CarrierID: carrierID}}
	passes := &passRepository{}
	service := NewRoutePass(NewRouteSlot(nil, nil, &passSlotStore{slot: slot}, approvals, time.Minute), passes)
	ctx := context.Background()

	if _, err := service.Issue(ctx, ownerID, slot.ID); !errors.Is(err, domain.ErrApprovalNotReady) {
		t.Fatalf("issue before ready: %v", err)
	}
	if _, err := service.Issue(ctx, carrierID, slot.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("participant issued a pass: %v", err)
	}
	if _, err := service.Get(ctx, strangerID, slot.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("stranger read pass status: %v", err)
	}
	if passes.issues != 0 {
		t.Fatalf("premature issue stored a pass: %d", passes.issues)
	}

	approvals.state.InspectorID = inspectorID
	status, err := service.Get(ctx, inspectorID, slot.ID)
	if err != nil || !status.Ready || status.Pass != nil {
		t.Fatalf("inspector should see ready but unissued: %+v %v", status, err)
	}
	issued, err := service.Issue(ctx, ownerID, slot.ID)
	if err != nil || issued.Pass == nil || !issued.Ready || passes.issues != 1 {
		t.Fatalf("owner failed to issue one pass: %+v %v", issued, err)
	}

	if _, err := service.Get(ctx, strangerID, slot.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("stranger read issued pass: %v", err)
	}
	approvals.state = domain.ApprovalState{}
	for _, actorID := range []uuid.UUID{ownerID, carrierID, inspectorID} {
		stored, err := service.Get(ctx, actorID, slot.ID)
		if err != nil || stored.Pass == nil || stored.Ready || !stored.Pass.IssuedAt.Equal(issued.Pass.IssuedAt) {
			t.Fatalf("actor %s lost issued pass after approval expiry: %+v %v", actorID, stored, err)
		}
	}
	again, err := service.Issue(ctx, ownerID, slot.ID)
	if err != nil || again.Pass == nil || !again.Pass.IssuedAt.Equal(issued.Pass.IssuedAt) || passes.issues != 1 {
		t.Fatalf("repeat issue did not return stored pass: %+v %v", again, err)
	}
	state, err := service.slots.Status(ctx, ownerID, slot.ID)
	if err != nil || state.Slot.Code != "" {
		t.Fatalf("status exposed secret slot code: %+v %v", state, err)
	}
}
