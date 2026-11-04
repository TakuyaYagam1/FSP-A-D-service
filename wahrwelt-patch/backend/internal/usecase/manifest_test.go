package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

type manifestRepositoryStub struct {
	manifest domain.Manifest
	revision domain.ManifestRevision
}

func (r *manifestRepositoryStub) CreateManifest(_ context.Context, manifest domain.Manifest, revision domain.ManifestRevision) (domain.Manifest, domain.ManifestRevision, error) {
	return manifest, revision, nil
}

func (r *manifestRepositoryStub) GetManifestByNumber(context.Context, string) (domain.Manifest, error) {
	return r.manifest, nil
}

func (r *manifestRepositoryStub) ListManifestRevisions(context.Context, uuid.UUID) ([]domain.RevisionSummary, error) {
	return nil, nil
}

func (r *manifestRepositoryStub) GetManifestRevision(context.Context, uuid.UUID, uuid.UUID) (domain.ManifestRevision, error) {
	return r.revision, nil
}

func TestRevisionRequiresManifestOwner(t *testing.T) {
	ownerID := uuid.New()
	manifestID := uuid.New()
	revisionID := uuid.New()
	repository := &manifestRepositoryStub{
		manifest: domain.Manifest{ID: manifestID, OwnerID: ownerID, Number: "MNF-" + manifestID.String()},
		revision: domain.ManifestRevision{ID: revisionID, ManifestID: manifestID, OwnerID: ownerID, SealedRoute: "north road", CreatedAt: time.Now()},
	}
	service := NewManifest(repository)

	_, err := service.Revision(context.Background(), uuid.New(), repository.manifest.Number, revisionID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Revision() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestRevisionIsAvailableToManifestOwner(t *testing.T) {
	ownerID := uuid.New()
	manifestID := uuid.New()
	revisionID := uuid.New()
	repository := &manifestRepositoryStub{
		manifest: domain.Manifest{ID: manifestID, OwnerID: ownerID, Number: "MNF-" + manifestID.String()},
		revision: domain.ManifestRevision{ID: revisionID, ManifestID: manifestID, OwnerID: ownerID, SealedRoute: "north road", CreatedAt: time.Now()},
	}
	service := NewManifest(repository)

	got, err := service.Revision(context.Background(), ownerID, repository.manifest.Number, revisionID)
	if err != nil {
		t.Fatalf("Revision() error = %v", err)
	}
	if got.SealedRoute != repository.revision.SealedRoute {
		t.Fatalf("Revision() SealedRoute = %q, want %q", got.SealedRoute, repository.revision.SealedRoute)
	}
}
