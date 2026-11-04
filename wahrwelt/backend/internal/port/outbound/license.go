package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

type LicenseRepository interface {
	CreateLicense(context.Context, domain.License) (domain.License, error)
	GetLicenseByNumber(context.Context, string) (domain.License, error)
}
