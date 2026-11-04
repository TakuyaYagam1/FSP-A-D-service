package outbound

import (
	"context"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

type RouteSlotRepository interface {
	CreateRouteSlot(context.Context, domain.RouteSlot) (domain.RouteSlot, error)
	GetRouteSlot(context.Context, uuid.UUID) (domain.RouteSlot, error)
}

type ApprovalStore interface {
	RecordApproval(context.Context, uuid.UUID, uuid.UUID, domain.ApprovalRole, time.Duration) (domain.ApprovalState, error)
	GetApprovals(context.Context, uuid.UUID) (domain.ApprovalState, error)
}
