package postgres

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Store) CreateContract(ctx context.Context, item domain.Contract) (domain.Contract, error) {
	row, err := s.queries.CreateContract(ctx, sqlc.CreateContractParams{
		ID:             item.ID,
		OwnerID:        item.OwnerID,
		ContractNumber: item.Number,
		CrewName:       item.CrewName,
		Service:        item.Service,
		CrewPhrase:     item.CrewPhrase,
		StartsAt:       pgtype.Timestamptz{Time: item.StartsAt, Valid: true},
		EndsAt:         pgtype.Timestamptz{Time: item.EndsAt, Valid: true},
	})
	if err != nil {
		return domain.Contract{}, translate(err)
	}
	return fromContract(row), nil
}

func (s *Store) GetContractByNumber(ctx context.Context, number string) (domain.Contract, error) {
	row, err := s.queries.GetContractByNumber(ctx, number)
	if err != nil {
		return domain.Contract{}, translate(err)
	}
	return fromContract(row), nil
}

func fromContract(row sqlc.ServiceContract) domain.Contract {
	return domain.Contract{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		Number:     row.ContractNumber,
		CrewName:   row.CrewName,
		Service:    row.Service,
		CrewPhrase: row.CrewPhrase,
		StartsAt:   row.StartsAt.Time,
		EndsAt:     row.EndsAt.Time,
		CreatedAt:  row.CreatedAt.Time,
	}
}
