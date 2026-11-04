package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"uuid"
)

const (
	maxQuotaVolume       = 1_000_000_000
	maxQuotaFields       = 4
	maxQuotaFieldLength  = 24
	maxQuotaFieldsLength = 128
)

type Quota struct {
	repository outbound.QuotaRepository
}

func NewQuota(repository outbound.QuotaRepository) *Quota {
	return &Quota{repository: repository}
}

func (q *Quota) Create(ctx context.Context, ownerID uuid.UUID, quotaCode string, volume int64, expiresAt time.Time) (domain.Quota, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.Quota{}, domain.ErrUnauthorized
	}
	quotaCode, err := domain.ValidateRouteText(quotaCode)
	if err != nil || len(quotaCode) > 120 || volume < 1 || volume > maxQuotaVolume || !expiresAt.After(time.Now()) {
		return domain.Quota{}, domain.ErrInvalidInput
	}
	id := uuid.New()
	item := domain.Quota{
		ID:        id,
		OwnerID:   ownerID,
		Number:    "QUO-" + id.String(),
		QuotaCode: quotaCode,
		Volume:    volume,
		ExpiresAt: expiresAt.UTC(),
		CreatedAt: time.Now().UTC(),
	}
	created, err := q.repository.CreateQuota(ctx, item)
	if err != nil {
		return domain.Quota{}, fmt.Errorf("create quota: %w", err)
	}
	return created, nil
}

func (q *Quota) Get(ctx context.Context, ownerID uuid.UUID, number string) (domain.Quota, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.Quota{}, domain.ErrUnauthorized
	}
	item, err := q.repository.GetQuotaByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.Quota{}, fmt.Errorf("get quota: %w", err)
	}
	if item.OwnerID != ownerID {
		return domain.Quota{}, domain.ErrNotFound
	}
	return item, nil
}

func (q *Quota) Summary(ctx context.Context, number, rawFields string) (domain.QuotaSummary, error) {
	fields, err := selectQuotaFields(rawFields)
	if err != nil {
		return domain.QuotaSummary{}, err
	}
	item, err := q.repository.GetQuotaByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.QuotaSummary{}, fmt.Errorf("get quota summary: %w", err)
	}
	return domain.QuotaSummary{Quota: item, Fields: fields}, nil
}

func selectQuotaFields(raw string) ([]string, error) {
	if raw == "" {
		return []string{"volume", "expires_at"}, nil
	}
	if len(raw) > maxQuotaFieldsLength {
		return nil, domain.ErrInvalidInput
	}
	identifiers := strings.Split(raw, ",")
	if len(identifiers) > maxQuotaFields {
		return nil, domain.ErrInvalidInput
	}
	fields := make([]string, 0, len(identifiers))
	seen := make(map[string]struct{}, len(identifiers))
	for _, identifier := range identifiers {
		if identifier == "quota_code" || !validQuotaIdentifier(identifier) {
			return nil, domain.ErrInvalidInput
		}
		field := strings.ToLower(identifier)
		if field != "quota_number" && field != "quota_code" && field != "volume" && field != "expires_at" {
			return nil, domain.ErrInvalidInput
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	return fields, nil
}

func validQuotaIdentifier(value string) bool {
	if len(value) == 0 || len(value) > maxQuotaFieldLength {
		return false
	}
	for index, char := range value {
		letter := (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
		if index == 0 {
			if !letter && char != '_' {
				return false
			}
			continue
		}
		if !letter && (char < '0' || char > '9') && char != '_' {
			return false
		}
	}
	return true
}
