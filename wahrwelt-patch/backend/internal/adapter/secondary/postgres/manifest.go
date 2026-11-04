package postgres

import (
	"context"
	"fmt"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

func (s *Store) CreateManifest(ctx context.Context, manifest domain.Manifest, revision domain.ManifestRevision) (domain.Manifest, domain.ManifestRevision, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Manifest{}, domain.ManifestRevision{}, fmt.Errorf("begin manifest transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := s.queries.WithTx(tx)
	manifestRow, err := queries.CreateManifest(ctx, sqlc.CreateManifestParams{
		ID:             manifest.ID,
		OwnerID:        manifest.OwnerID,
		ManifestNumber: manifest.Number,
		Origin:         manifest.Origin,
		Destination:    manifest.Destination,
		Status:         string(manifest.Status),
	})
	if err != nil {
		return domain.Manifest{}, domain.ManifestRevision{}, translate(err)
	}
	revisionRow, err := queries.CreateManifestRevision(ctx, sqlc.CreateManifestRevisionParams{
		ID:          revision.ID,
		ManifestID:  revision.ManifestID,
		OwnerID:     revision.OwnerID,
		Status:      string(revision.Status),
		SealedRoute: revision.SealedRoute,
	})
	if err != nil {
		return domain.Manifest{}, domain.ManifestRevision{}, translate(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Manifest{}, domain.ManifestRevision{}, fmt.Errorf("commit manifest transaction: %w", err)
	}
	created := fromManifest(manifestRow)
	createdRevision := fromManifestRevision(revisionRow)
	createdRevision.ManifestNumber = created.Number
	return created, createdRevision, nil
}

func (s *Store) GetManifestByNumber(ctx context.Context, number string) (domain.Manifest, error) {
	row, err := s.queries.GetManifestByNumber(ctx, number)
	if err != nil {
		return domain.Manifest{}, translate(err)
	}
	return fromManifest(row), nil
}

func (s *Store) ListManifestRevisions(ctx context.Context, manifestID uuid.UUID) ([]domain.RevisionSummary, error) {
	rows, err := s.queries.ListManifestRevisions(ctx, manifestID)
	if err != nil {
		return nil, translate(err)
	}
	result := make([]domain.RevisionSummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.RevisionSummary{
			ID:        row.ID,
			Status:    domain.ManifestStatus(row.Status),
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return result, nil
}

func (s *Store) GetManifestRevision(ctx context.Context, manifestID, revisionID uuid.UUID) (domain.ManifestRevision, error) {
	row, err := s.queries.GetManifestRevision(ctx, sqlc.GetManifestRevisionParams{ManifestID: manifestID, RevisionID: revisionID})
	if err != nil {
		return domain.ManifestRevision{}, translate(err)
	}
	return domain.ManifestRevision{
		ID:             row.ID,
		ManifestID:     row.ManifestID,
		OwnerID:        row.OwnerID,
		ManifestNumber: row.ManifestNumber,
		Status:         domain.ManifestStatus(row.Status),
		SealedRoute:    row.SealedRoute,
		CreatedAt:      row.CreatedAt.Time,
	}, nil
}

func fromManifest(row sqlc.Manifest) domain.Manifest {
	return domain.Manifest{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		Number:      row.ManifestNumber,
		Status:      domain.ManifestStatus(row.Status),
		Origin:      row.Origin,
		Destination: row.Destination,
		CreatedAt:   row.CreatedAt.Time,
	}
}

func fromManifestRevision(row sqlc.ManifestRevision) domain.ManifestRevision {
	return domain.ManifestRevision{
		ID:          row.ID,
		ManifestID:  row.ManifestID,
		OwnerID:     row.OwnerID,
		Status:      domain.ManifestStatus(row.Status),
		SealedRoute: row.SealedRoute,
		CreatedAt:   row.CreatedAt.Time,
	}
}
