package postgres

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (s *Store) CreateGatePass(ctx context.Context, item domain.GatePass) (domain.GatePass, error) {
	row, err := s.queries.CreateGatePass(ctx, sqlc.CreateGatePassParams{
		ID:             item.ID,
		OwnerID:        item.OwnerID,
		GateOperatorID: item.GateOperatorID,
		GatePassNumber: item.Number,
		State:          item.State,
		GateWord:       item.GateWord,
	})
	if err != nil {
		return domain.GatePass{}, translate(err)
	}
	return fromGatePass(row), nil
}

func (s *Store) GetGatePassByNumber(ctx context.Context, number string) (domain.GatePass, error) {
	row, err := s.queries.GetGatePassByNumber(ctx, number)
	if err != nil {
		return domain.GatePass{}, translate(err)
	}
	return fromGatePass(row), nil
}

func fromGatePass(row sqlc.GatePass) domain.GatePass {
	return domain.GatePass{
		ID:             row.ID,
		OwnerID:        row.OwnerID,
		GateOperatorID: row.GateOperatorID,
		Number:         row.GatePassNumber,
		State:          row.State,
		GateWord:       row.GateWord,
		CreatedAt:      row.CreatedAt.Time,
	}
}
