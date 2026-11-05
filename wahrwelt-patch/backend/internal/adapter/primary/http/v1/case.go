package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (h *Handler) CreateCase(ctx context.Context, req *api.CreateCaseRequest) (*api.Case, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.cases.Create(ctx, current.UserID, string(req.Kind), req.Origin, req.Destination)
	if err != nil {
		return nil, err
	}
	return apiCase(item), nil
}

func (h *Handler) ListCases(ctx context.Context, params api.ListCasesParams) (*api.CaseList, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	items, more, err := h.cases.List(ctx, current.UserID, params.Offset.Or(0))
	if err != nil {
		return nil, err
	}
	result := &api.CaseList{Cases: make([]api.OwnedCase, 0, len(items)), HasMore: more}
	for _, item := range items {
		statuses, err := h.routeSlots.List(ctx, current.UserID, item.ID)
		if err != nil {
			return nil, err
		}
		entry := api.OwnedCase{Case: *apiCase(item), RouteSlots: make([]api.RouteSlotStatus, 0, len(statuses))}
		for _, status := range statuses {
			entry.RouteSlots = append(entry.RouteSlots, apiRouteSlotStatus(status))
		}
		result.Cases = append(result.Cases, entry)
	}
	return result, nil
}

func (h *Handler) GetCase(ctx context.Context, params api.GetCaseParams) (*api.Case, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	caseID, err := parseID(params.CaseID)
	if err != nil {
		return nil, err
	}
	item, err := h.cases.Get(ctx, current.UserID, caseID)
	if err != nil {
		return nil, err
	}
	return apiCase(item), nil
}

func apiCase(item domain.Case) *api.Case {
	return &api.Case{
		CaseID:      item.ID.String(),
		OwnerID:     item.OwnerID.String(),
		Kind:        api.CaseKind(item.Kind),
		Origin:      item.Origin,
		Destination: item.Destination,
		CreatedAt:   item.CreatedAt,
	}
}
