package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (h *Handler) CreateLicense(ctx context.Context, req *api.CreateLicenseRequest) (*api.LicenseCreated, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.License.Create(ctx, current.UserID, req.ReleasePin)
	if err != nil {
		return nil, err
	}
	return &api.LicenseCreated{
		LicenseNumber: item.Number,
		Status:        api.LicenseCreatedStatus(item.Status),
		CreatedAt:     item.CreatedAt,
	}, nil
}

func (h *Handler) GetLicenseCard(ctx context.Context, params api.GetLicenseCardParams) (*api.LicenseCard, error) {
	item, err := h.features.License.PublicCard(ctx, params.LicenseNumber)
	if err != nil {
		return nil, err
	}
	return &api.LicenseCard{LicenseNumber: item.Number, Status: api.LicenseCardStatus(item.Status)}, nil
}

func (h *Handler) GetLicense(ctx context.Context, params api.GetLicenseParams) (*api.License, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.License.Get(ctx, current.UserID, params.LicenseNumber)
	if err != nil {
		return nil, err
	}
	return apiLicense(item), nil
}

func (h *Handler) IssueLicenseToken(ctx context.Context, req *api.LicenseTokenRequest, params api.IssueLicenseTokenParams) (*api.LicenseTokenHeaders, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	token, err := h.features.License.IssueToken(ctx, current.UserID, params.LicenseNumber, string(req.Scope))
	if err != nil {
		return nil, err
	}
	return &api.LicenseTokenHeaders{
		CacheControl: "no-store",
		Response: api.LicenseToken{
			Token:     token.Value,
			ExpiresAt: token.ExpiresAt,
		},
	}, nil
}

func (h *Handler) GetLicenseReleasePin(ctx context.Context, params api.GetLicenseReleasePinParams) (*api.LicenseReleasePin, error) {
	token, err := middleware.CurrentLicenseToken(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.License.ReadWithToken(ctx, token, params.LicenseNumber)
	if err != nil {
		return nil, err
	}
	return &api.LicenseReleasePin{LicenseNumber: item.Number, ReleasePin: item.ReleasePin}, nil
}

func apiLicense(item domain.License) *api.License {
	return &api.License{
		LicenseNumber: item.Number,
		Status:        api.LicenseStatus(item.Status),
		ReleasePin:    item.ReleasePin,
		CreatedAt:     item.CreatedAt,
	}
}
