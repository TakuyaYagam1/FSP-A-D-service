package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

type quotaRepositoryStub struct {
	item domain.Quota
}

func (r *quotaRepositoryStub) CreateQuota(_ context.Context, item domain.Quota) (domain.Quota, error) {
	r.item = item
	return item, nil
}

func (r *quotaRepositoryStub) GetQuotaByNumber(_ context.Context, _ string) (domain.Quota, error) {
	return r.item, nil
}

func TestSummaryRejectsClosedFieldAfterNormalization(t *testing.T) {
	repository := &quotaRepositoryStub{item: domain.Quota{
		Number:    "QUO-1",
		QuotaCode: "closed-code",
		Volume:    42,
		ExpiresAt: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
	}}
	service := NewQuota(repository)

	_, err := service.Summary(context.Background(), repository.item.Number, "QUOTA_CODE")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("Summary() error = %v, want invalid input", err)
	}
}

func TestSummaryNormalizesPublicFieldNames(t *testing.T) {
	repository := &quotaRepositoryStub{item: domain.Quota{
		Number:    "QUO-1",
		QuotaCode: "closed-code",
		Volume:    42,
		ExpiresAt: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
	}}
	service := NewQuota(repository)

	got, err := service.Summary(context.Background(), repository.item.Number, "VOLUME,EXPIRES_AT,QUOTA_NUMBER")
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	want := []string{"volume", "expires_at", "quota_number"}
	if !reflect.DeepEqual(got.Fields, want) {
		t.Fatalf("Summary() fields = %v, want %v", got.Fields, want)
	}
}
