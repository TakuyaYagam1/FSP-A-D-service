package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

type RoutePassRepository interface {
	GetRoutePass(context.Context, uuid.UUID) (domain.RoutePass, error)
	CreateRoutePass(context.Context, uuid.UUID) (domain.RoutePass, error)
}
