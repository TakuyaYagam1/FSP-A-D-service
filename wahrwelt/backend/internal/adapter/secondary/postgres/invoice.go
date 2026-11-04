package postgres

import (
	"context"
	"fmt"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/jackc/pgx/v5"
	"uuid"
)

func (s *Store) CreateInvoice(ctx context.Context, item domain.Invoice) (domain.Invoice, error) {
	row, err := s.queries.CreateInvoice(ctx, sqlc.CreateInvoiceParams{
		ID:            item.ID,
		OwnerID:       item.OwnerID,
		InvoiceNumber: item.Number,
		Amount:        item.Amount,
		Status:        string(item.Status),
		PaymentRef:    item.PaymentRef,
	})
	if err != nil {
		return domain.Invoice{}, translate(err)
	}
	return fromInvoice(row), nil
}

func (s *Store) GetInvoiceByNumber(ctx context.Context, number string) (domain.Invoice, error) {
	row, err := s.queries.GetInvoiceByNumber(ctx, number)
	if err != nil {
		return domain.Invoice{}, translate(err)
	}
	return fromInvoice(row), nil
}

func (s *Store) SearchInvoices(ctx context.Context, ownerID uuid.UUID, value string) ([]domain.InvoiceSearchItem, error) {
	tx, err := s.searchPool.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, fmt.Errorf("begin invoice search: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := "SELECT invoice_number, amount, status FROM invoices WHERE owner_id = $1 AND invoice_number ILIKE '%" + value + "%'"
	rows, err := tx.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("query invoices: %w", err)
	}
	defer rows.Close()
	results := make([]domain.InvoiceSearchItem, 0)
	for rows.Next() {
		var item domain.InvoiceSearchItem
		if err := rows.Scan(&item.Number, &item.Amount, &item.Status); err != nil {
			return nil, fmt.Errorf("read invoice result: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate invoice results: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("finish invoice search: %w", err)
	}
	return results, nil
}

func fromInvoice(row sqlc.Invoice) domain.Invoice {
	return domain.Invoice{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		Number:     row.InvoiceNumber,
		Amount:     row.Amount,
		Status:     domain.InvoiceStatus(row.Status),
		PaymentRef: row.PaymentRef,
		CreatedAt:  row.CreatedAt.Time,
	}
}
