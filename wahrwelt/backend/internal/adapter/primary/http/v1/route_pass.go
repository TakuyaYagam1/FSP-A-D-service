package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (h *Handler) GetRoutePass(ctx context.Context, params api.GetRoutePassParams) (*api.RoutePassState, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	slotID, err := parseID(params.SlotID)
	if err != nil {
		return nil, err
	}
	state, err := h.features.RoutePass.Get(ctx, current.UserID, slotID)
	if err != nil {
		return nil, err
	}
	return apiRoutePassState(state), nil
}

func (h *Handler) IssueRoutePass(ctx context.Context, params api.IssueRoutePassParams) (*api.RoutePassState, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	slotID, err := parseID(params.SlotID)
	if err != nil {
		return nil, err
	}
	state, err := h.features.RoutePass.Issue(ctx, current.UserID, slotID)
	if err != nil {
		return nil, err
	}
	return apiRoutePassState(state), nil
}

func apiRoutePassState(state domain.RoutePassState) *api.RoutePassState {
	response := &api.RoutePassState{Ready: state.Ready, Issued: state.Pass != nil}
	if state.Pass != nil {
		response.Pass = api.NewOptRoutePass(api.RoutePass{
			RouteSlotID:       state.Pass.SlotID.String(),
			CaseID:            state.Pass.CaseID.String(),
			Kind:              api.RoutePassKind(state.Pass.Kind),
			Origin:            state.Pass.Origin,
			Destination:       state.Pass.Destination,
			CarrierUsername:   state.Pass.CarrierUsername,
			InspectorUsername: state.Pass.InspectorUsername,
			IssuedAt:          state.Pass.IssuedAt,
		})
	}
	return response
}
