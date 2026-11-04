package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

type InspectionRepository interface {
	CreateInspection(context.Context, domain.Inspection) (domain.Inspection, error)
	GetInspectionByNumber(context.Context, string) (domain.Inspection, error)
	ListInspectionComments(context.Context, uuid.UUID) ([]domain.InspectionComment, error)
	AddInspectionComment(context.Context, domain.InspectionComment, *domain.InspectionVisibility) (domain.InspectionComment, error)
}
