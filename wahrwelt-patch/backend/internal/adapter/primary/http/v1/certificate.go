package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (h *Handler) CreateCertificate(ctx context.Context, req *api.CreateCertificateRequest) (*api.Certificate, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	certificate, err := h.features.Certificate.Create(ctx, current.UserID, req.OriginPhrase)
	if err != nil {
		return nil, err
	}
	return apiCertificate(certificate), nil
}

func (h *Handler) GetCertificate(ctx context.Context, params api.GetCertificateParams) (*api.Certificate, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	certificate, err := h.features.Certificate.Get(ctx, current.UserID, params.Serial)
	if err != nil {
		return nil, err
	}
	return apiCertificate(certificate), nil
}

func (h *Handler) VerifyCertificate(ctx context.Context, req *api.VerifyCertificateRequest) (*api.VerifyCertificateResponse, error) {
	body, err := h.features.Certificate.Verify(ctx, req.Serial, req.RegistryURL.String())
	if err != nil {
		return nil, err
	}
	return &api.VerifyCertificateResponse{Serial: req.Serial, Body: body}, nil
}

func (h *Handler) GetInternalCertificateProof(ctx context.Context, params api.GetInternalCertificateProofParams) (*api.CertificateProof, error) {
	phrase, err := h.features.Certificate.Proof(ctx, params.Serial)
	if err != nil {
		return nil, err
	}
	return &api.CertificateProof{OriginPhrase: phrase}, nil
}

func apiCertificate(certificate domain.CertificateSummary) *api.Certificate {
	return &api.Certificate{Serial: certificate.Serial, OriginPhrase: certificate.OriginPhrase, CreatedAt: certificate.CreatedAt}
}
