package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

const maxPreviewPNG = 4 << 20
const maxPreviewLog = 8 << 10

type Preview struct {
	root string
}

func NewPreview(root string) *Preview {
	return &Preview{root: root}
}

func (p *Preview) CreatePreview(ctx context.Context, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	directory, err := p.previewPath(id)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(p.root, 0750); err != nil {
		return fmt.Errorf("create preview directory: %w", err)
	}
	if err := os.Mkdir(directory, 0700); err != nil {
		if os.IsExist(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("create preview job directory: %w", err)
	}
	if err := writeAtomic(filepath.Join(directory, "state"), []byte("created\n"), 0600); err != nil {
		_ = os.RemoveAll(directory)
		return fmt.Errorf("initialize preview job: %w", err)
	}
	return nil
}

func (p *Preview) RemovePreview(ctx context.Context, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	directory, err := p.previewPath(id)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(directory); err != nil {
		return fmt.Errorf("remove preview job: %w", err)
	}
	return nil
}

func (p *Preview) QueuePreview(ctx context.Context, id uuid.UUID, input []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(input) == 0 || len(input) > 1<<20 || !strings.HasPrefix(string(input), "%!") {
		return domain.ErrInvalidInput
	}
	directory, err := p.previewPath(id)
	if err != nil {
		return err
	}
	inputPath := filepath.Join(directory, "input.eps")
	file, err := os.OpenFile(inputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			return domain.ErrConflict
		}
		if os.IsNotExist(err) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("open preview input: %w", err)
	}
	if _, err := file.Write(input); err != nil {
		_ = file.Close()
		_ = os.Remove(inputPath)
		return fmt.Errorf("write preview input: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(inputPath)
		return fmt.Errorf("sync preview input: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(inputPath)
		return fmt.Errorf("close preview input: %w", err)
	}
	if err := writeAtomic(filepath.Join(directory, "state"), []byte("queued\n"), 0600); err != nil {
		_ = os.Remove(inputPath)
		return fmt.Errorf("update preview state: %w", err)
	}
	if err := writeAtomic(filepath.Join(directory, "request"), []byte("ready\n"), 0600); err != nil {
		_ = os.Remove(inputPath)
		return fmt.Errorf("queue preview: %w", err)
	}
	return nil
}

func (p *Preview) GetPreviewResult(ctx context.Context, id uuid.UUID) (domain.PreviewResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.PreviewResult{}, err
	}
	directory, err := p.previewPath(id)
	if err != nil {
		return domain.PreviewResult{}, err
	}
	stateData, err := readBounded(filepath.Join(directory, "state"), 32)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.PreviewResult{}, domain.ErrNotFound
		}
		return domain.PreviewResult{}, fmt.Errorf("read preview state: %w", err)
	}
	state := strings.TrimSpace(string(stateData))
	if state != "created" && state != "queued" && state != "processing" && state != "complete" && state != "failed" {
		return domain.PreviewResult{}, fmt.Errorf("invalid preview state")
	}
	result := domain.PreviewResult{ID: id, State: state}
	if state != "complete" && state != "failed" {
		return result, nil
	}
	result.PNG, err = readBounded(filepath.Join(directory, "result.png"), maxPreviewPNG)
	if err != nil && !os.IsNotExist(err) {
		return domain.PreviewResult{}, fmt.Errorf("read preview image: %w", err)
	}
	result.Stdout, err = readBoundedText(filepath.Join(directory, "stdout.txt"), maxPreviewLog)
	if err != nil && !os.IsNotExist(err) {
		return domain.PreviewResult{}, fmt.Errorf("read preview output: %w", err)
	}
	result.Stderr, err = readBoundedText(filepath.Join(directory, "stderr.txt"), maxPreviewLog)
	if err != nil && !os.IsNotExist(err) {
		return domain.PreviewResult{}, fmt.Errorf("read preview errors: %w", err)
	}
	return result, nil
}

func (p *Preview) previewPath(id uuid.UUID) (string, error) {
	if id == (uuid.UUID{}) {
		return "", domain.ErrInvalidInput
	}
	return filepath.Join(p.root, id.String()), nil
}

func readBounded(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, limit+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if int64(len(data)) > limit {
		return nil, errors.New("preview file exceeds limit")
	}
	return data, nil
}

func readBoundedText(path string, limit int64) (string, error) {
	data, err := readBounded(path, limit)
	return string(data), err
}
