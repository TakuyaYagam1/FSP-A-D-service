package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

type InvoiceRepository interface {
	CreateInvoice(context.Context, domain.Invoice) (domain.Invoice, error)
	GetInvoiceByNumber(context.Context, string) (domain.Invoice, error)
	SearchInvoices(context.Context, uuid.UUID, string) ([]domain.InvoiceSearchItem, error)
}
