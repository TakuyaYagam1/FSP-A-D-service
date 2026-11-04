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
