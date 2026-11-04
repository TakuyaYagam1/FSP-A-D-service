package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSearchTreatsQueryAsData(t *testing.T) {
	dsn := os.Getenv("WAHRWELT_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("WAHRWELT_TEST_POSTGRES_DSN is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer admin.Close()

	schema := "invoice_test_" + strings.ReplaceAll(uuid.New().String(), "-", "")
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	defer func() {
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+quotedSchema+" CASCADE")
	}()

	_, err = admin.Exec(ctx, fmt.Sprintf(`CREATE TABLE %s.invoices (
		owner_id uuid NOT NULL,
		invoice_number text NOT NULL,
		amount bigint NOT NULL,
		status text NOT NULL,
		payment_ref text NOT NULL
	)`, quotedSchema))
	if err != nil {
		t.Fatalf("create test invoice table: %v", err)
	}
	ownerID := uuid.New()
	invoiceNumber := "INV-" + uuid.New().String()
	paymentRef := "payment-ref-private"
	if _, err := admin.Exec(ctx,
		fmt.Sprintf("INSERT INTO %s.invoices (owner_id, invoice_number, amount, status, payment_ref) VALUES ($1, $2, $3, $4, $5)", quotedSchema),
		ownerID, invoiceNumber, int64(125), "pending", paymentRef,
	); err != nil {
		t.Fatalf("insert test invoice: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse test database DSN: %v", err)
	}
	poolConfig.MaxConns = 1
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schema
	searchPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("open search pool: %v", err)
	}
	defer searchPool.Close()

	store := &Store{searchPool: &InvoiceSearchPool{pool: searchPool}}
	normal, err := store.SearchInvoices(ctx, ownerID, invoiceNumber)
	if err != nil {
		t.Fatalf("SearchInvoices() for an owned invoice error = %v", err)
	}
	if len(normal) != 1 || normal[0].Number != invoiceNumber || normal[0].Amount != 125 || normal[0].Status != "pending" {
		t.Fatalf("SearchInvoices() for an owned invoice = %#v", normal)
	}

	query := fmt.Sprintf("missing%%' UNION SELECT invoice_number, amount, payment_ref FROM invoices WHERE invoice_number = '%s'--", invoiceNumber)
	results, err := store.SearchInvoices(ctx, uuid.New(), query)
	if err != nil {
		t.Fatalf("SearchInvoices() error = %v", err)
	}
	for _, item := range results {
		if item.Status == paymentRef {
			t.Fatalf("SearchInvoices() returned private data in status for %s", item.Number)
		}
	}
}
