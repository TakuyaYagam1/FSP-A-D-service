package postgres

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (s *Store) CreateLicense(ctx context.Context, item domain.License) (domain.License, error) {
	row, err := s.queries.CreateLicense(ctx, sqlc.CreateLicenseParams{
		ID:            item.ID,
		OwnerID:       item.OwnerID,
		LicenseNumber: item.Number,
		Status:        string(item.Status),
		ReleasePin:    item.ReleasePin,
	})
	if err != nil {
		return domain.License{}, translate(err)
	}
	return fromLicense(row), nil
}

func (s *Store) GetLicenseByNumber(ctx context.Context, number string) (domain.License, error) {
	row, err := s.queries.GetLicenseByNumber(ctx, number)
	if err != nil {
		return domain.License{}, translate(err)
	}
	return fromLicense(row), nil
}

func fromLicense(row sqlc.ExportLicense) domain.License {
	return domain.License{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		Number:     row.LicenseNumber,
		Status:     domain.LicenseStatus(row.Status),
		ReleasePin: row.ReleasePin,
		CreatedAt:  row.CreatedAt.Time,
	}
}
