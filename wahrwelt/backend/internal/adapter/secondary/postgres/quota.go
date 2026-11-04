package postgres

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Store) CreateQuota(ctx context.Context, item domain.Quota) (domain.Quota, error) {
	row, err := s.queries.CreateQuota(ctx, sqlc.CreateQuotaParams{
		ID:          item.ID,
		OwnerID:     item.OwnerID,
		QuotaNumber: item.Number,
		QuotaCode:   item.QuotaCode,
		Volume:      item.Volume,
		ExpiresAt:   pgtype.Timestamptz{Time: item.ExpiresAt, Valid: true},
	})
	if err != nil {
		return domain.Quota{}, translate(err)
	}
	return fromQuota(row), nil
}

func (s *Store) GetQuotaByNumber(ctx context.Context, number string) (domain.Quota, error) {
	row, err := s.queries.GetQuotaByNumber(ctx, number)
	if err != nil {
		return domain.Quota{}, translate(err)
	}
	return fromQuota(row), nil
}

func fromQuota(row sqlc.Quota) domain.Quota {
	return domain.Quota{
		ID:        row.ID,
		OwnerID:   row.OwnerID,
		Number:    row.QuotaNumber,
		QuotaCode: row.QuotaCode,
		Volume:    row.Volume,
		ExpiresAt: row.ExpiresAt.Time,
		CreatedAt: row.CreatedAt.Time,
	}
}
