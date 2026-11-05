package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

type CaseRepository interface {
	CreateCase(context.Context, domain.Case) (domain.Case, error)
	GetCase(context.Context, uuid.UUID) (domain.Case, error)
	ListCases(context.Context, uuid.UUID, int32) ([]domain.Case, error)
}
