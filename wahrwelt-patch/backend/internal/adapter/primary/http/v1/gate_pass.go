package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (h *Handler) CreateGatePass(ctx context.Context, req *api.CreateGatePassRequest) (*api.GatePassCreated, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.GatePass.Create(ctx, current.UserID, req.GateOperatorUsername, string(req.State), req.GateWord)
	if err != nil {
		return nil, err
	}
	return &api.GatePassCreated{
		GatePassNumber: item.Number,
		State:          api.GatePassCreatedState(item.State),
		CreatedAt:      item.CreatedAt,
	}, nil
}

func (h *Handler) GetGatePass(ctx context.Context, params api.GetGatePassParams) (*api.GatePass, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.GatePass.Get(ctx, current.UserID, params.GatePassNumber)
	if err != nil {
		return nil, err
	}
	return apiGatePass(item), nil
}

func (h *Handler) GetGatePassCard(ctx context.Context, params api.GetGatePassCardParams) (*api.GatePassCard, error) {
	item, err := h.features.GatePass.PublicCard(ctx, params.GatePassNumber)
	if err != nil {
		return nil, err
	}
	return &api.GatePassCard{GatePassNumber: item.Number, State: api.GatePassCardState(item.State)}, nil
}

func (h *Handler) GetGatePassSealed(ctx context.Context, params api.GetGatePassSealedParams) (*api.GatePass, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	// Заголовок роли остается в контракте, но не определяет права пользователя
	item, err := h.features.GatePass.GetSealed(ctx, current.UserID, params.GatePassNumber)
	if err != nil {
		return nil, err
	}
	return apiGatePass(item), nil
}

func apiGatePass(item domain.GatePass) *api.GatePass {
	return &api.GatePass{
		GatePassNumber: item.Number,
		State:          api.GatePassState(item.State),
		GateWord:       item.GateWord,
		CreatedAt:      item.CreatedAt,
	}
}
