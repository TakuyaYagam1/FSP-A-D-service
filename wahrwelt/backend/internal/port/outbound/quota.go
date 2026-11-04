package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

type QuotaRepository interface {
	CreateQuota(context.Context, domain.Quota) (domain.Quota, error)
	GetQuotaByNumber(context.Context, string) (domain.Quota, error)
}
