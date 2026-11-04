package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (h *Handler) CreateContract(ctx context.Context, req *api.CreateContractRequest) (*api.ContractCreated, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.Contract.Create(ctx, current.UserID, req.CrewName, req.Service, req.CrewPhrase, req.StartsAt, req.EndsAt)
	if err != nil {
		return nil, err
	}
	return &api.ContractCreated{ContractNumber: item.Number, CreatedAt: item.CreatedAt}, nil
}

func (h *Handler) GetContract(ctx context.Context, params api.GetContractParams) (*api.Contract, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.Contract.Get(ctx, current.UserID, params.ContractNumber)
	if err != nil {
		return nil, err
	}
	return apiContract(item), nil
}

func (h *Handler) GetContractCard(ctx context.Context, params api.GetContractCardParams) (*api.ContractCard, error) {
	item, err := h.features.Contract.PublicCard(ctx, params.ContractNumber)
	if err != nil {
		return nil, err
	}
	return &api.ContractCard{
		ContractNumber: item.Number,
		CrewName:       item.CrewName,
		Service:        item.Service,
		StartsAt:       item.StartsAt,
		EndsAt:         item.EndsAt,
	}, nil
}

func (h *Handler) PreviewContract(ctx context.Context, req *api.ContractPreviewRequest, params api.PreviewContractParams) (*api.ContractPreview, error) {
	preview, err := h.features.Contract.Preview(ctx, params.ContractNumber, req.Template)
	if err != nil {
		return nil, err
	}
	return &api.ContractPreview{Rendered: preview.Rendered}, nil
}

func apiContract(item domain.Contract) *api.Contract {
	return &api.Contract{
		ContractNumber: item.Number,
		CrewName:       item.CrewName,
		Service:        item.Service,
		CrewPhrase:     item.CrewPhrase,
		StartsAt:       item.StartsAt,
		EndsAt:         item.EndsAt,
		CreatedAt:      item.CreatedAt,
	}
}
