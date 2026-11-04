package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"uuid"
)

type Manifest struct{ repository outbound.ManifestRepository }

func NewManifest(repository outbound.ManifestRepository) *Manifest {
	return &Manifest{repository: repository}
}

func (m *Manifest) Create(ctx context.Context, ownerID uuid.UUID, origin, destination, sealedRoute string) (domain.Manifest, domain.ManifestRevision, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.Manifest{}, domain.ManifestRevision{}, domain.ErrUnauthorized
	}
	origin, err := domain.ValidateRouteText(origin)
	if err != nil {
		return domain.Manifest{}, domain.ManifestRevision{}, err
	}
	destination, err = domain.ValidateRouteText(destination)
	if err != nil {
		return domain.Manifest{}, domain.ManifestRevision{}, err
	}
	sealedRoute = strings.TrimSpace(sealedRoute)
	if sealedRoute == "" || len(sealedRoute) > 240 {
		return domain.Manifest{}, domain.ManifestRevision{}, domain.ErrInvalidInput
	}

	manifestID := uuid.New()
	revision := domain.ManifestRevision{
		ID:          uuid.New(),
		ManifestID:  manifestID,
		OwnerID:     ownerID,
		Status:      domain.ManifestStatusReceived,
		SealedRoute: sealedRoute,
		CreatedAt:   time.Now().UTC(),
	}
	manifest := domain.Manifest{
		ID:          manifestID,
		OwnerID:     ownerID,
		Number:      "MNF-" + manifestID.String(),
		Status:      domain.ManifestStatusReceived,
		Origin:      origin,
		Destination: destination,
		CreatedAt:   revision.CreatedAt,
	}
	revision.ManifestNumber = manifest.Number
	created, createdRevision, err := m.repository.CreateManifest(ctx, manifest, revision)
	if err != nil {
		return domain.Manifest{}, domain.ManifestRevision{}, fmt.Errorf("create manifest: %w", err)
	}
	return created, createdRevision, nil
}

func (m *Manifest) PublicCard(ctx context.Context, number string) (domain.Manifest, error) {
	manifest, err := m.repository.GetManifestByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.Manifest{}, fmt.Errorf("get manifest: %w", err)
	}
	return manifest, nil
}

func (m *Manifest) History(ctx context.Context, number string) (domain.Manifest, []domain.RevisionSummary, error) {
	manifest, err := m.repository.GetManifestByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.Manifest{}, nil, fmt.Errorf("get manifest: %w", err)
	}
	revisions, err := m.repository.ListManifestRevisions(ctx, manifest.ID)
	if err != nil {
		return domain.Manifest{}, nil, fmt.Errorf("list manifest revisions: %w", err)
	}
	return manifest, revisions, nil
}

func (m *Manifest) Revision(ctx context.Context, actorID uuid.UUID, number string, revisionID uuid.UUID) (domain.ManifestRevision, error) {
	if actorID == (uuid.UUID{}) {
		return domain.ManifestRevision{}, domain.ErrUnauthorized
	}
	manifest, err := m.repository.GetManifestByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.ManifestRevision{}, fmt.Errorf("get manifest: %w", err)
	}
	revision, err := m.repository.GetManifestRevision(ctx, manifest.ID, revisionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ManifestRevision{}, domain.ErrNotFound
		}
		return domain.ManifestRevision{}, fmt.Errorf("get manifest revision: %w", err)
	}
	return revision, nil
}
