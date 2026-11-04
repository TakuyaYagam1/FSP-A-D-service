package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

type ContractRepository interface {
	CreateContract(context.Context, domain.Contract) (domain.Contract, error)
	GetContractByNumber(context.Context, string) (domain.Contract, error)
}
