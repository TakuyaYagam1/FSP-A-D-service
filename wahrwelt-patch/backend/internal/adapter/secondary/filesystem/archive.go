package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"uuid"
)

type Archive struct {
	root string
}

func NewArchive(root string) *Archive {
	return &Archive{root: root}
}

func (a *Archive) WriteWaybill(ctx context.Context, number, code string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := a.waybillPath(number)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(a.root, 0750); err != nil {
		return fmt.Errorf("create archive directory: %w", err)
	}
	if err := writeAtomic(path, []byte(strings.TrimSpace(code)+"\n"), 0440); err != nil {
		return fmt.Errorf("write archive record: %w", err)
	}
	return nil
}

func (a *Archive) RemoveWaybill(ctx context.Context, number string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := a.waybillPath(number)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove archive record: %w", err)
	}
	return nil
}

func (a *Archive) waybillPath(number string) (string, error) {
	if !strings.HasPrefix(number, "AWB-") {
		return "", fmt.Errorf("invalid waybill number")
	}
	if _, err := uuid.Parse(strings.TrimPrefix(number, "AWB-")); err != nil {
		return "", fmt.Errorf("invalid waybill number")
	}
	return filepath.Join(a.root, number+".txt"), nil
}
