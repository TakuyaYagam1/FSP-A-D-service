package usecase

import (
	"context"
	"testing"

	"uuid"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

type certificateRepositoryStub struct {
	item domain.Certificate
}

func (r *certificateRepositoryStub) CreateCertificate(_ context.Context, item domain.Certificate) (domain.Certificate, error) {
	return item, nil
}

func (r *certificateRepositoryStub) GetCertificateBySerial(_ context.Context, serial string) (domain.Certificate, error) {
	if serial != r.item.Serial {
		return domain.Certificate{}, domain.ErrNotFound
	}
	return r.item, nil
}

func TestGetReturnsOriginPhraseToOwner(t *testing.T) {
	ownerID := uuid.New()
	item := domain.Certificate{
		OwnerID:      ownerID,
		Serial:       "CRT-01234567-89ab-cdef-0123-456789abcdef",
		OriginPhrase: "stored origin phrase",
	}
	service := NewCertificate(&certificateRepositoryStub{item: item}, nil)

	got, err := service.Get(context.Background(), ownerID, item.Serial)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.OriginPhrase != item.OriginPhrase {
		t.Fatalf("Get() OriginPhrase = %q, want %q", got.OriginPhrase, item.OriginPhrase)
	}
}
