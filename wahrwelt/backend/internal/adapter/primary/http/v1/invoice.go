package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (h *Handler) CreateInvoice(ctx context.Context, req *api.CreateInvoiceRequest) (*api.InvoiceSummary, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.Invoice.Create(ctx, current.UserID, req.Amount, req.PaymentRef)
	if err != nil {
		return nil, err
	}
	return apiInvoiceSummary(item), nil
}

func (h *Handler) GetInvoice(ctx context.Context, params api.GetInvoiceParams) (*api.Invoice, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	item, err := h.features.Invoice.Get(ctx, current.UserID, params.InvoiceNumber)
	if err != nil {
		return nil, err
	}
	return &api.Invoice{
		InvoiceNumber: item.Number,
		Amount:        item.Amount,
		Status:        api.InvoiceStatus(item.Status),
		PaymentRef:    item.PaymentRef,
		CreatedAt:     item.CreatedAt,
	}, nil
}

func (h *Handler) SearchInvoices(ctx context.Context, params api.SearchInvoicesParams) (*api.InvoiceSearchResponse, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.features.Invoice.Search(ctx, current.UserID, params.Q)
	if err != nil {
		return nil, err
	}
	results := make([]api.InvoiceSearchItem, 0, len(items))
	for _, item := range items {
		results = append(results, api.InvoiceSearchItem{InvoiceNumber: item.Number, Amount: item.Amount, Status: item.Status})
	}
	return &api.InvoiceSearchResponse{Results: results}, nil
}

func apiInvoiceSummary(item domain.InvoiceSummary) *api.InvoiceSummary {
	return &api.InvoiceSummary{
		InvoiceNumber: item.Number,
		Amount:        item.Amount,
		Status:        api.InvoiceSummaryStatus(item.Status),
		CreatedAt:     item.CreatedAt,
	}
}
