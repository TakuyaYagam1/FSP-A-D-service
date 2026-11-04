package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/TakuyaYagam1/wahrwelt/config"
	"github.com/TakuyaYagam1/wahrwelt/db"
	"github.com/TakuyaYagam1/wahrwelt/internal/bootstrap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if action, ok, err := migrationAction(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	} else if ok {
		if err := db.Migrate(ctx, action, os.Getenv("POSTGRES_DSN")); err != nil {
			fmt.Fprintln(os.Stderr, "migration failed")
			os.Exit(1)
		}
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid service configuration")
		os.Exit(1)
	}
	if err := runService(ctx,
		func(ctx context.Context) error { return db.Migrate(ctx, "up", cfg.PostgresDSN) },
		func(ctx context.Context) error { return bootstrap.Run(ctx, cfg) },
	); err != nil {
		fmt.Fprintln(os.Stderr, "service stopped after an error")
		os.Exit(1)
	}
}

func migrationAction(args []string) (string, bool, error) {
	if len(args) == 0 {
		return "", false, nil
	}
	if len(args) != 1 {
		return "", false, errors.New("usage: wahrwelt [migrate|migrate-down]")
	}
	switch args[0] {
	case "migrate":
		return "up", true, nil
	case "migrate-down":
		return "down", true, nil
	default:
		return "", false, errors.New("usage: wahrwelt [migrate|migrate-down]")
	}
}

func runService(ctx context.Context, migrate, serve func(context.Context) error) error {
	if err := migrate(ctx); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return serve(ctx)
}
