package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (h *Handler) CreateQuota(ctx context.Context, req *api.CreateQuotaRequest) (*api.QuotaCreated, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.Quota.Create(ctx, current.UserID, req.QuotaCode, req.Volume, req.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &api.QuotaCreated{QuotaNumber: item.Number, CreatedAt: item.CreatedAt}, nil
}

func (h *Handler) GetQuota(ctx context.Context, params api.GetQuotaParams) (*api.Quota, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.Quota.Get(ctx, current.UserID, params.QuotaNumber)
	if err != nil {
		return nil, err
	}
	return apiQuota(item), nil
}

func (h *Handler) GetQuotaSummary(ctx context.Context, params api.GetQuotaSummaryParams) (*api.QuotaSummary, error) {
	item, err := h.features.Quota.Summary(ctx, params.QuotaNumber, params.Fields.Or(""))
	if err != nil {
		return nil, err
	}
	response := &api.QuotaSummary{}
	// quota_code не попадает в публичный ответ даже при наличии поля в схеме
	for _, field := range item.Fields {
		switch field {
		case "quota_number":
			response.SetQuotaNumber(api.NewOptString(item.Number))
		case "volume":
			response.SetVolume(api.NewOptInt64(item.Volume))
		case "expires_at":
			response.SetExpiresAt(api.NewOptDateTime(item.ExpiresAt))
		}
	}
	return response, nil
}

func apiQuota(item domain.Quota) *api.Quota {
	return &api.Quota{
		QuotaNumber: item.Number,
		QuotaCode:   item.QuotaCode,
		Volume:      item.Volume,
		ExpiresAt:   item.ExpiresAt,
		CreatedAt:   item.CreatedAt,
	}
}
