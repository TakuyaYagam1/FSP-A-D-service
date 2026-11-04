package v1

import (
	"context"
	"encoding/base64"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
)

func (h *Handler) CreatePreview(ctx context.Context) (*api.PreviewCreated, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.Preview.Create(ctx, current.UserID)
	if err != nil {
		return nil, err
	}
	return &api.PreviewCreated{PreviewID: item.ID.String(), State: api.PreviewCreatedStateCreated}, nil
}

func (h *Handler) UploadPreview(ctx context.Context, req *api.PreviewUploadRequest, params api.UploadPreviewParams) (*api.PreviewAccepted, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseID(params.PreviewID)
	if err != nil {
		return nil, err
	}
	if err := h.features.Preview.Upload(ctx, current.UserID, id, req.EpsBase64); err != nil {
		return nil, err
	}
	return &api.PreviewAccepted{PreviewID: id.String(), State: api.PreviewAcceptedStateQueued}, nil
}

func (h *Handler) GetPreview(ctx context.Context, params api.GetPreviewParams) (*api.Preview, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseID(params.PreviewID)
	if err != nil {
		return nil, err
	}
	item, err := h.features.Preview.Get(ctx, current.UserID, id)
	if err != nil {
		return nil, err
	}
	response := &api.Preview{PreviewID: item.ID.String(), State: api.PreviewState(item.State)}
	if len(item.PNG) != 0 {
		response.PNGBase64 = api.NewOptString(base64.StdEncoding.EncodeToString(item.PNG))
	}
	if item.Stdout != "" {
		response.Stdout = api.NewOptString(item.Stdout)
	}
	if item.Stderr != "" {
		response.Stderr = api.NewOptString(item.Stderr)
	}
	return response, nil
}
