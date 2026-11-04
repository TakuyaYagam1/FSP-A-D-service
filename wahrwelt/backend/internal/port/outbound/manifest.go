package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

type ManifestRepository interface {
	CreateManifest(context.Context, domain.Manifest, domain.ManifestRevision) (domain.Manifest, domain.ManifestRevision, error)
	GetManifestByNumber(context.Context, string) (domain.Manifest, error)
	ListManifestRevisions(context.Context, uuid.UUID) ([]domain.RevisionSummary, error)
	GetManifestRevision(context.Context, uuid.UUID, uuid.UUID) (domain.ManifestRevision, error)
}
