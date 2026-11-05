package postgres

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

func (s *Store) CreateCase(ctx context.Context, item domain.Case) (domain.Case, error) {
	row, err := s.queries.CreateCase(ctx, sqlc.CreateCaseParams{
		ID:          item.ID,
		OwnerID:     item.OwnerID,
		Kind:        string(item.Kind),
		Origin:      item.Origin,
		Destination: item.Destination,
	})
	if err != nil {
		return domain.Case{}, translate(err)
	}
	return fromCase(row), nil
}

func (s *Store) GetCase(ctx context.Context, id uuid.UUID) (domain.Case, error) {
	row, err := s.queries.GetCase(ctx, id)
	if err != nil {
		return domain.Case{}, translate(err)
	}
	return fromCase(row), nil
}

func (s *Store) ListCases(ctx context.Context, ownerID uuid.UUID, offset int32) ([]domain.Case, error) {
	rows, err := s.queries.ListCasesByOwner(ctx, sqlc.ListCasesByOwnerParams{OwnerID: ownerID, PageOffset: offset})
	if err != nil {
		return nil, translate(err)
	}
	items := make([]domain.Case, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromCase(row))
	}
	return items, nil
}

func fromCase(row sqlc.Case) domain.Case {
	return domain.Case{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		Kind:        domain.CaseKind(row.Kind),
		Origin:      row.Origin,
		Destination: row.Destination,
		CreatedAt:   row.CreatedAt.Time,
	}
}
