package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

type contractRepositoryStub struct {
	item domain.Contract
}

func (r *contractRepositoryStub) CreateContract(_ context.Context, item domain.Contract) (domain.Contract, error) {
	r.item = item
	return item, nil
}

func (r *contractRepositoryStub) GetContractByNumber(_ context.Context, _ string) (domain.Contract, error) {
	return r.item, nil
}

func TestPreviewDoesNotExposeCrewPhrase(t *testing.T) {
	contract := domain.Contract{
		Number:     "CTR-1",
		CrewName:   "Repair crew",
		Service:    "Pump repair",
		CrewPhrase: "closed phrase",
		StartsAt:   time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC),
		EndsAt:     time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC),
	}
	service := NewContract(&contractRepositoryStub{item: contract})

	preview, err := service.Preview(context.Background(), contract.Number, "{{.CrewPhrase}}")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("Preview() error = %v, want invalid input (rendered %q)", err, preview.Rendered)
	}
	if preview.Rendered != "" {
		t.Fatalf("Preview() rendered closed data %q", preview.Rendered)
	}
}

func TestPreviewRendersPublicContractFields(t *testing.T) {
	contract := domain.Contract{Number: "CTR-1", CrewName: "Repair crew", Service: "Pump repair"}
	service := NewContract(&contractRepositoryStub{item: contract})

	preview, err := service.Preview(context.Background(), contract.Number, "{{.CrewName}} / {{.Service}}")
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	if preview.Rendered != "Repair crew / Pump repair" {
		t.Fatalf("Preview() rendered = %q", preview.Rendered)
	}
}

func TestPreviewDotContainsOnlyPublicFields(t *testing.T) {
	contract := domain.Contract{
		Number:     "CTR-1",
		CrewName:   "Repair crew",
		Service:    "Pump repair",
		CrewPhrase: "closed phrase",
	}
	service := NewContract(&contractRepositoryStub{item: contract})

	preview, err := service.Preview(context.Background(), contract.Number, "{{.}}")
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	if strings.Contains(preview.Rendered, contract.CrewPhrase) {
		t.Fatalf("Preview() rendered closed phrase: %q", preview.Rendered)
	}
	if !strings.Contains(preview.Rendered, contract.CrewName) || !strings.Contains(preview.Rendered, contract.Service) {
		t.Fatalf("Preview() omitted public fields: %q", preview.Rendered)
	}
}
