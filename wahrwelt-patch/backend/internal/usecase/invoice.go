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

type Invoice struct {
	repository outbound.InvoiceRepository
}

func NewInvoice(repository outbound.InvoiceRepository) *Invoice {
	return &Invoice{repository: repository}
}

func (i *Invoice) Create(ctx context.Context, ownerID uuid.UUID, amount int64, paymentRef string) (domain.InvoiceSummary, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.InvoiceSummary{}, domain.ErrUnauthorized
	}
	paymentRef = strings.TrimSpace(paymentRef)
	if amount <= 0 || paymentRef == "" || len(paymentRef) > 240 {
		return domain.InvoiceSummary{}, domain.ErrInvalidInput
	}
	id := uuid.New()
	item := domain.Invoice{
		ID:         id,
		OwnerID:    ownerID,
		Number:     "INV-" + id.String(),
		Amount:     amount,
		Status:     domain.InvoiceStatusPending,
		PaymentRef: paymentRef,
		CreatedAt:  time.Now().UTC(),
	}
	created, err := i.repository.CreateInvoice(ctx, item)
	if err != nil {
		return domain.InvoiceSummary{}, fmt.Errorf("create invoice: %w", err)
	}
	return invoiceSummary(created), nil
}

func (i *Invoice) Get(ctx context.Context, ownerID uuid.UUID, number string) (domain.Invoice, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.Invoice{}, domain.ErrUnauthorized
	}
	item, err := i.repository.GetInvoiceByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.Invoice{}, fmt.Errorf("get invoice: %w", err)
	}
	if item.OwnerID != ownerID {
		return domain.Invoice{}, domain.ErrNotFound
	}
	return item, nil
}

func (i *Invoice) Search(ctx context.Context, ownerID uuid.UUID, query string) ([]domain.InvoiceSearchItem, error) {
	if ownerID == (uuid.UUID{}) {
		return nil, domain.ErrUnauthorized
	}
	query = strings.TrimSpace(query)
	if query == "" || len(query) > 120 {
		return nil, domain.ErrInvalidInput
	}
	results, err := i.repository.SearchInvoices(ctx, ownerID, query)
	if err != nil {
		return nil, fmt.Errorf("search invoices: %w", err)
	}
	return results, nil
}

func invoiceSummary(item domain.Invoice) domain.InvoiceSummary {
	return domain.InvoiceSummary{Number: item.Number, Amount: item.Amount, Status: item.Status, CreatedAt: item.CreatedAt}
}
