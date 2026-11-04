package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

type CertificateRepository interface {
	CreateCertificate(context.Context, domain.Certificate) (domain.Certificate, error)
	GetCertificateBySerial(context.Context, string) (domain.Certificate, error)
}

type CertificateVerifier interface {
	Get(context.Context, string) (string, error)
}
