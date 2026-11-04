package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

type licenseRepositoryStub struct {
	items map[string]domain.License
}

func (r *licenseRepositoryStub) CreateLicense(_ context.Context, item domain.License) (domain.License, error) {
	r.items[item.Number] = item
	return item, nil
}

func (r *licenseRepositoryStub) GetLicenseByNumber(_ context.Context, number string) (domain.License, error) {
	item, ok := r.items[number]
	if !ok {
		return domain.License{}, domain.ErrNotFound
	}
	return item, nil
}

func TestReadWithTokenBindsLicenseAndOwner(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	owned := domain.License{OwnerID: ownerID, Number: "LIC-owned", ReleasePin: "owned-pin"}
	other := domain.License{OwnerID: otherOwnerID, Number: "LIC-other", ReleasePin: "other-pin"}
	repository := &licenseRepositoryStub{items: map[string]domain.License{
		owned.Number: owned,
		other.Number: other,
	}}
	licenses := NewLicense(repository, "test-secret", time.Minute)
	token, err := licenses.IssueToken(ctx, ownerID, owned.Number, LicenseReadScope)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	got, err := licenses.ReadWithToken(ctx, token.Value, other.Number)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("ReadWithToken() for another license error = %v, want unauthorized (got %#v)", err, got)
	}

	owned.OwnerID = otherOwnerID
	repository.items[owned.Number] = owned
	got, err = licenses.ReadWithToken(ctx, token.Value, owned.Number)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("ReadWithToken() after owner change error = %v, want unauthorized (got %#v)", err, got)
	}
}

func TestReadWithTokenAllowsIssuedLicense(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	item := domain.License{OwnerID: ownerID, Number: "LIC-owned", ReleasePin: "owned-pin"}
	licenses := NewLicense(&licenseRepositoryStub{items: map[string]domain.License{item.Number: item}}, "test-secret", time.Minute)
	token, err := licenses.IssueToken(ctx, ownerID, item.Number, LicenseReadScope)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	got, err := licenses.ReadWithToken(ctx, token.Value, item.Number)
	if err != nil {
		t.Fatalf("ReadWithToken() error = %v", err)
	}
	if got.ReleasePin != item.ReleasePin {
		t.Fatalf("ReadWithToken() release pin = %q, want %q", got.ReleasePin, item.ReleasePin)
	}
}
