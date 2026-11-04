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
