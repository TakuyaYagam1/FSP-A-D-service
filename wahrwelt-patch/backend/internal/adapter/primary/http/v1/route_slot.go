package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (h *Handler) CreateRouteSlot(ctx context.Context, req *api.CreateRouteSlotRequest, params api.CreateRouteSlotParams) (*api.RouteSlot, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	caseID, err := parseID(params.CaseID)
	if err != nil {
		return nil, err
	}
	item, err := h.routeSlots.Create(ctx, current.UserID, caseID, req.CarrierUsername, req.InspectorUsername, req.SlotCode)
	if err != nil {
		return nil, err
	}
	return apiRouteSlot(item), nil
}

func (h *Handler) ListCaseRouteSlots(ctx context.Context, params api.ListCaseRouteSlotsParams) (*api.RouteSlotStatusList, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	caseID, err := parseID(params.CaseID)
	if err != nil {
		return nil, err
	}
	statuses, err := h.routeSlots.List(ctx, current.UserID, caseID)
	if err != nil {
		return nil, err
	}
	result := &api.RouteSlotStatusList{RouteSlots: make([]api.RouteSlotStatus, 0, len(statuses))}
	for _, status := range statuses {
		result.RouteSlots = append(result.RouteSlots, apiRouteSlotStatus(status))
	}
	return result, nil
}

func apiRouteSlotStatus(status domain.RouteSlotStatus) api.RouteSlotStatus {
	return api.RouteSlotStatus{
		Slot: *apiRouteSlot(status.Slot), ApprovalCount: status.ApprovalCount, Ready: status.Ready,
	}
}

func (h *Handler) ApproveRouteSlot(ctx context.Context, req *api.ApprovalRequest, params api.ApproveRouteSlotParams) (*api.ApprovalState, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	slotID, err := parseID(params.SlotID)
	if err != nil {
		return nil, err
	}
	result, err := h.routeSlots.Approve(ctx, current.UserID, slotID, domain.ApprovalRole(req.Role))
	if err != nil {
		return nil, err
	}
	return &api.ApprovalState{
		CarrierApproved:   result.CarrierApproved,
		InspectorApproved: result.InspectorApproved,
		Ready:             result.Ready,
	}, nil
}

func (h *Handler) GetRouteSlotCode(ctx context.Context, params api.GetRouteSlotCodeParams) (*api.RouteSlotCode, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	slotID, err := parseID(params.SlotID)
	if err != nil {
		return nil, err
	}
	code, err := h.routeSlots.Code(ctx, current.UserID, slotID)
	if err != nil {
		return nil, err
	}
	return &api.RouteSlotCode{SlotCode: code}, nil
}

func apiRouteSlot(item domain.RouteSlot) *api.RouteSlot {
	return &api.RouteSlot{
		RouteSlotID:       item.ID.String(),
		CaseID:            item.CaseID.String(),
		CarrierUsername:   item.CarrierUsername,
		InspectorUsername: item.InspectorUsername,
		CreatedAt:         item.CreatedAt,
	}
}
