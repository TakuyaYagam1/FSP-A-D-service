package postgres

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (s *Store) CreateCertificate(ctx context.Context, item domain.Certificate) (domain.Certificate, error) {
	row, err := s.queries.CreateCertificate(ctx, sqlc.CreateCertificateParams{
		ID:           item.ID,
		OwnerID:      item.OwnerID,
		Serial:       item.Serial,
		OriginPhrase: item.OriginPhrase,
	})
	if err != nil {
		return domain.Certificate{}, translate(err)
	}
	return fromCertificate(row), nil
}

func (s *Store) GetCertificateBySerial(ctx context.Context, serial string) (domain.Certificate, error) {
	row, err := s.queries.GetCertificateBySerial(ctx, serial)
	if err != nil {
		return domain.Certificate{}, translate(err)
	}
	return fromCertificate(row), nil
}

func fromCertificate(row sqlc.Certificate) domain.Certificate {
	return domain.Certificate{
		ID:           row.ID,
		OwnerID:      row.OwnerID,
		Serial:       row.Serial,
		OriginPhrase: row.OriginPhrase,
		CreatedAt:    row.CreatedAt.Time,
	}
}
