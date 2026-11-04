package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"uuid"
)

const maxPreviewBytes = 1 << 20

type Preview struct {
	repository outbound.PreviewRepository
	files      outbound.PreviewFiles
}

func NewPreview(repository outbound.PreviewRepository, files outbound.PreviewFiles) *Preview {
	return &Preview{repository: repository, files: files}
}

func (p *Preview) Create(ctx context.Context, ownerID uuid.UUID) (domain.Preview, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.Preview{}, domain.ErrUnauthorized
	}
	item := domain.Preview{ID: uuid.New(), OwnerID: ownerID}
	if err := p.files.CreatePreview(ctx, item.ID); err != nil {
		return domain.Preview{}, fmt.Errorf("create preview files: %w", err)
	}
	created, err := p.repository.CreatePreview(ctx, item)
	if err != nil {
		_ = p.files.RemovePreview(ctx, item.ID)
		return domain.Preview{}, fmt.Errorf("create preview: %w", err)
	}
	return created, nil
}

func (p *Preview) Upload(ctx context.Context, ownerID uuid.UUID, id uuid.UUID, encoded string) error {
	if ownerID == (uuid.UUID{}) {
		return domain.ErrUnauthorized
	}
	if len(encoded) == 0 || len(encoded) > base64.StdEncoding.EncodedLen(maxPreviewBytes) {
		return domain.ErrInvalidInput
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(data) == 0 || len(data) > maxPreviewBytes || !strings.HasPrefix(string(data), "%!") {
		return domain.ErrInvalidInput
	}
	if _, err := p.repository.GetPreviewForOwner(ctx, id, ownerID); err != nil {
		return fmt.Errorf("get preview for upload: %w", err)
	}
	if err := p.files.QueuePreview(ctx, id, data); err != nil {
		return fmt.Errorf("queue preview: %w", err)
	}
	return nil
}

func (p *Preview) Get(ctx context.Context, ownerID uuid.UUID, id uuid.UUID) (domain.PreviewResult, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.PreviewResult{}, domain.ErrUnauthorized
	}
	if _, err := p.repository.GetPreviewForOwner(ctx, id, ownerID); err != nil {
		return domain.PreviewResult{}, fmt.Errorf("get preview owner: %w", err)
	}
	result, err := p.files.GetPreviewResult(ctx, id)
	if err != nil {
		return domain.PreviewResult{}, fmt.Errorf("get preview result: %w", err)
	}
	return result, nil
}
