package bootstrap

import (
	"context"
	"fmt"

	"github.com/TakuyaYagam1/wahrwelt/config"
)

func Run(ctx context.Context, cfg config.Config) error {
	app, cleanup, err := initializeApp(ctx, cfg)
	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}
	defer cleanup()
	return app.Run(ctx)
}
