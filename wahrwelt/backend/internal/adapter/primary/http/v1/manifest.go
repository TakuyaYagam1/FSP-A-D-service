package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
)

func (h *Handler) CreateManifest(ctx context.Context, req *api.CreateManifestRequest) (*api.ManifestCreated, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	manifest, revision, err := h.features.Manifest.Create(ctx, current.UserID, req.Origin, req.Destination, req.SealedRoute)
	if err != nil {
		return nil, err
	}
	return &api.ManifestCreated{
		ManifestNumber: manifest.Number,
		RevisionID:     revision.ID.String(),
		Status:         api.ManifestCreatedStatus(manifest.Status),
	}, nil
}

func (h *Handler) GetManifest(ctx context.Context, params api.GetManifestParams) (*api.ManifestCard, error) {
	manifest, err := h.features.Manifest.PublicCard(ctx, params.ManifestNumber)
	if err != nil {
		return nil, err
	}
	return &api.ManifestCard{ManifestNumber: manifest.Number, Status: api.ManifestCardStatus(manifest.Status)}, nil
}

func (h *Handler) ListManifestRevisions(ctx context.Context, params api.ListManifestRevisionsParams) (*api.ManifestHistory, error) {
	manifest, revisions, err := h.features.Manifest.History(ctx, params.ManifestNumber)
	if err != nil {
		return nil, err
	}
	items := make([]api.RevisionSummary, 0, len(revisions))
	for _, revision := range revisions {
		items = append(items, api.RevisionSummary{
			RevisionID: revision.ID.String(),
			Status:     api.RevisionSummaryStatus(revision.Status),
		})
	}
	return &api.ManifestHistory{ManifestNumber: manifest.Number, Revisions: items}, nil
}

func (h *Handler) GetManifestRevision(ctx context.Context, params api.GetManifestRevisionParams) (*api.ManifestRevision, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	revisionID, err := parseID(params.RevisionID)
	if err != nil {
		return nil, err
	}
	revision, err := h.features.Manifest.Revision(ctx, current.UserID, params.ManifestNumber, revisionID)
	if err != nil {
		return nil, err
	}
	return &api.ManifestRevision{
		ManifestNumber: revision.ManifestNumber,
		RevisionID:     revision.ID.String(),
		Status:         api.ManifestRevisionStatus(revision.Status),
		SealedRoute:    revision.SealedRoute,
		CreatedAt:      revision.CreatedAt,
	}, nil
}
