package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (h *Handler) CreateWaybill(ctx context.Context, req *api.CreateWaybillRequest) (*api.WaybillCreated, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.Waybill.Create(ctx, current.UserID, req.ArchiveCode)
	if err != nil {
		return nil, err
	}
	return &api.WaybillCreated{WaybillNumber: item.Number, State: api.WaybillCreatedStateReady, CreatedAt: item.CreatedAt}, nil
}

func (h *Handler) GetWaybill(ctx context.Context, params api.GetWaybillParams) (*api.Waybill, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.Waybill.Get(ctx, current.UserID, params.WaybillNumber)
	if err != nil {
		return nil, err
	}
	return apiWaybill(item), nil
}

func (h *Handler) GetWaybillCard(ctx context.Context, params api.GetWaybillCardParams) (*api.WaybillCard, error) {
	item, err := h.features.Waybill.PublicCard(ctx, params.WaybillNumber)
	if err != nil {
		return nil, err
	}
	return &api.WaybillCard{WaybillNumber: item.Number, State: api.WaybillCardStateReady}, nil
}

func apiWaybill(item domain.Waybill) *api.Waybill {
	return &api.Waybill{
		WaybillNumber: item.Number,
		ArchiveCode:   item.ArchiveCode,
		State:         api.WaybillStateReady,
		CreatedAt:     item.CreatedAt,
	}
}
