package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/config"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool       *pgxpool.Pool
	searchPool *InvoiceSearchPool
	queries    *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool, searchPool *InvoiceSearchPool) *Store {
	return &Store{pool: pool, searchPool: searchPool, queries: sqlc.New(pool)}
}

func NewPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, func(), error) {
	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("open PostgreSQL pool: %w", err)
	}
	cleanup := func() { pool.Close() }

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(connectCtx); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	return pool, cleanup, nil
}

type InvoiceSearchPool struct {
	pool *pgxpool.Pool
}

func NewInvoiceSearchPool(ctx context.Context, cfg config.Config) (*InvoiceSearchPool, func(), error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.InvoiceSearchDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("parse invoice search PostgreSQL DSN: %w", err)
	}
	poolConfig.MaxConns = 4
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("open invoice search PostgreSQL pool: %w", err)
	}
	cleanup := func() { pool.Close() }
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(connectCtx); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("connect invoice search PostgreSQL pool: %w", err)
	}
	return &InvoiceSearchPool{pool: pool}, cleanup, nil
}

func translate(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			return domain.ErrConflict
		case "23503":
			return domain.ErrNotFound
		case "23514", "22001":
			return domain.ErrInvalidInput
		}
	}
	return fmt.Errorf("PostgreSQL operation: %w", err)
}
