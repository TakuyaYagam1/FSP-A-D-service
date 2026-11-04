package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

type PreviewRepository interface {
	CreatePreview(context.Context, domain.Preview) (domain.Preview, error)
	GetPreviewForOwner(context.Context, uuid.UUID, uuid.UUID) (domain.Preview, error)
}

type PreviewFiles interface {
	CreatePreview(context.Context, uuid.UUID) error
	RemovePreview(context.Context, uuid.UUID) error
	QueuePreview(context.Context, uuid.UUID, []byte) error
	GetPreviewResult(context.Context, uuid.UUID) (domain.PreviewResult, error)
}
