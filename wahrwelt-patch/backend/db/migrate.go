package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var errInvalidAction = errors.New("unsupported migration action")

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(ctx context.Context, action, dsn string) error {
	if action != "up" && action != "down" {
		return fmt.Errorf("%w: %s", errInvalidAction, action)
	}
	if strings.TrimSpace(dsn) == "" {
		return errors.New("POSTGRES_DSN is required")
	}
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open PostgreSQL: %w", err)
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	files, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		_ = database.Close()
		return fmt.Errorf("open embedded migrations: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, database, files)
	if err != nil {
		_ = database.Close()
		return fmt.Errorf("configure PostgreSQL migrations: %w", err)
	}
	defer func() { _ = provider.Close() }()
	if action == "up" {
		_, err = provider.Up(ctx)
	} else {
		_, err = provider.Down(ctx)
	}
	if err != nil {
		return fmt.Errorf("run PostgreSQL migrations: %w", err)
	}
	return nil
}
