package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

type GatePassRepository interface {
	CreateGatePass(context.Context, domain.GatePass) (domain.GatePass, error)
	GetGatePassByNumber(context.Context, string) (domain.GatePass, error)
}
