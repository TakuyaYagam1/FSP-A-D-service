package postgres

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

func (s *Store) CreateWaybill(ctx context.Context, item domain.Waybill) (domain.Waybill, error) {
	row, err := s.queries.CreateWaybill(ctx, sqlc.CreateWaybillParams{
		ID:            item.ID,
		OwnerID:       item.OwnerID,
		WaybillNumber: item.Number,
		ArchiveCode:   item.ArchiveCode,
	})
	if err != nil {
		return domain.Waybill{}, translate(err)
	}
	return fromWaybill(row), nil
}

func (s *Store) GetWaybillByNumber(ctx context.Context, number string) (domain.Waybill, error) {
	row, err := s.queries.GetWaybillByNumber(ctx, number)
	if err != nil {
		return domain.Waybill{}, translate(err)
	}
	return fromWaybill(row), nil
}

func fromWaybill(row sqlc.ArchiveWaybill) domain.Waybill {
	return domain.Waybill{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		Number:      row.WaybillNumber,
		ArchiveCode: row.ArchiveCode,
		State:       row.State,
		CreatedAt:   row.CreatedAt.Time,
	}
}

func (s *Store) CreatePreview(ctx context.Context, item domain.Preview) (domain.Preview, error) {
	row, err := s.queries.CreatePreview(ctx, sqlc.CreatePreviewParams{ID: item.ID, OwnerID: item.OwnerID})
	if err != nil {
		return domain.Preview{}, translate(err)
	}
	return domain.Preview{ID: row.ID, OwnerID: row.OwnerID, CreatedAt: row.CreatedAt.Time}, nil
}

func (s *Store) GetPreviewForOwner(ctx context.Context, id, ownerID uuid.UUID) (domain.Preview, error) {
	row, err := s.queries.GetPreviewForOwner(ctx, sqlc.GetPreviewForOwnerParams{ID: id, OwnerID: ownerID})
	if err != nil {
		return domain.Preview{}, translate(err)
	}
	return domain.Preview{ID: row.ID, OwnerID: row.OwnerID, CreatedAt: row.CreatedAt.Time}, nil
}
