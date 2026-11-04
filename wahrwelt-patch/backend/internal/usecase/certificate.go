package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"uuid"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
)

type Certificate struct {
	repository outbound.CertificateRepository
	verifier   outbound.CertificateVerifier
}

func NewCertificate(repository outbound.CertificateRepository, verifier outbound.CertificateVerifier) *Certificate {
	return &Certificate{repository: repository, verifier: verifier}
}

func (c *Certificate) Create(ctx context.Context, ownerID uuid.UUID, phrase string) (domain.CertificateSummary, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.CertificateSummary{}, domain.ErrUnauthorized
	}
	phrase = strings.TrimSpace(phrase)
	if phrase == "" || len(phrase) > 240 {
		return domain.CertificateSummary{}, domain.ErrInvalidInput
	}
	id := uuid.New()
	item := domain.Certificate{
		ID:           id,
		OwnerID:      ownerID,
		Serial:       "CRT-" + id.String(),
		OriginPhrase: phrase,
		CreatedAt:    time.Now().UTC(),
	}
	created, err := c.repository.CreateCertificate(ctx, item)
	if err != nil {
		return domain.CertificateSummary{}, fmt.Errorf("create certificate: %w", err)
	}
	return certificateSummary(created), nil
}

func (c *Certificate) Get(ctx context.Context, ownerID uuid.UUID, serial string) (domain.CertificateSummary, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.CertificateSummary{}, domain.ErrUnauthorized
	}
	item, err := c.repository.GetCertificateBySerial(ctx, strings.TrimSpace(serial))
	if err != nil {
		return domain.CertificateSummary{}, fmt.Errorf("get certificate: %w", err)
	}
	if item.OwnerID != ownerID {
		return domain.CertificateSummary{}, domain.ErrNotFound
	}
	return certificateSummary(item), nil
}

func (c *Certificate) Proof(ctx context.Context, serial string) (string, error) {
	item, err := c.repository.GetCertificateBySerial(ctx, strings.TrimSpace(serial))
	if err != nil {
		return "", fmt.Errorf("get certificate proof: %w", err)
	}
	return item.OriginPhrase, nil
}

func (c *Certificate) Verify(ctx context.Context, serial, targetURL string) (string, error) {
	if _, err := c.repository.GetCertificateBySerial(ctx, strings.TrimSpace(serial)); err != nil {
		return "", fmt.Errorf("find certificate: %w", err)
	}
	body, err := c.verifier.Get(ctx, targetURL)
	if err != nil {
		return "", fmt.Errorf("verify certificate registry: %w", err)
	}
	return body, nil
}

func certificateSummary(item domain.Certificate) domain.CertificateSummary {
	return domain.CertificateSummary{ID: item.ID, OwnerID: item.OwnerID, Serial: item.Serial, OriginPhrase: item.OriginPhrase, CreatedAt: item.CreatedAt}
}
