package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"uuid"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
)

type Waybill struct {
	repository outbound.WaybillRepository
	files      outbound.WaybillFiles
}

func NewWaybill(repository outbound.WaybillRepository, files outbound.WaybillFiles) *Waybill {
	return &Waybill{repository: repository, files: files}
}

func (w *Waybill) Create(ctx context.Context, ownerID uuid.UUID, archiveCode string) (domain.Waybill, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.Waybill{}, domain.ErrUnauthorized
	}
	archiveCode = strings.TrimSpace(archiveCode)
	if archiveCode == "" || len(archiveCode) > 240 {
		return domain.Waybill{}, domain.ErrInvalidInput
	}
	id := uuid.New()
	item := domain.Waybill{
		ID:          id,
		OwnerID:     ownerID,
		Number:      "AWB-" + id.String(),
		ArchiveCode: archiveCode,
		State:       "ready",
		CreatedAt:   time.Now().UTC(),
	}
	if err := w.files.WriteWaybill(ctx, item.Number, item.ArchiveCode); err != nil {
		return domain.Waybill{}, fmt.Errorf("write waybill record: %w", err)
	}
	created, err := w.repository.CreateWaybill(ctx, item)
	if err != nil {
		_ = w.files.RemoveWaybill(ctx, item.Number)
		return domain.Waybill{}, fmt.Errorf("create waybill: %w", err)
	}
	return created, nil
}

func (w *Waybill) Get(ctx context.Context, ownerID uuid.UUID, number string) (domain.Waybill, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.Waybill{}, domain.ErrUnauthorized
	}
	item, err := w.repository.GetWaybillByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.Waybill{}, fmt.Errorf("get waybill: %w", err)
	}
	if item.OwnerID != ownerID {
		return domain.Waybill{}, domain.ErrNotFound
	}
	return item, nil
}

func (w *Waybill) PublicCard(ctx context.Context, number string) (domain.WaybillCard, error) {
	item, err := w.repository.GetWaybillByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.WaybillCard{}, fmt.Errorf("get waybill card: %w", err)
	}
	return domain.WaybillCard{Number: item.Number, State: item.State}, nil
}
