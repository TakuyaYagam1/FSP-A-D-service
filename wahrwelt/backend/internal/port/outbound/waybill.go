package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

type WaybillRepository interface {
	CreateWaybill(context.Context, domain.Waybill) (domain.Waybill, error)
	GetWaybillByNumber(context.Context, string) (domain.Waybill, error)
}

type WaybillFiles interface {
	WriteWaybill(context.Context, string, string) error
	RemoveWaybill(context.Context, string) error
}
