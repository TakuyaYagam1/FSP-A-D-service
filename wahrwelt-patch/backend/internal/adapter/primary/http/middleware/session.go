package middleware

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/usecase"
	"uuid"
)

type SecurityHandler struct {
	auth     *usecase.Auth
	licenses *usecase.License
}

type Principal struct {
	UserID uuid.UUID
	Token  string
}

type principalKey struct{}
type licenseTokenKey struct{}

func NewSecurityHandler(auth *usecase.Auth, licenses *usecase.License) *SecurityHandler {
	return &SecurityHandler{auth: auth, licenses: licenses}
}

func (s *SecurityHandler) HandleSessionCookie(ctx context.Context, _ api.OperationName, cookie api.SessionCookie) (context.Context, error) {
	userID, err := s.auth.ResolveSession(ctx, cookie.APIKey)
	if err != nil {
		return nil, err
	}
	current := Principal{UserID: userID, Token: cookie.APIKey}
	return context.WithValue(ctx, principalKey{}, current), nil
}

func (s *SecurityHandler) HandleLicenseReadToken(ctx context.Context, _ api.OperationName, token api.LicenseReadToken) (context.Context, error) {
	if s.licenses == nil || s.licenses.ValidateReadToken(token.Token) != nil {
		return nil, domain.ErrUnauthorized
	}
	return context.WithValue(ctx, licenseTokenKey{}, token.Token), nil
}

func CurrentPrincipal(ctx context.Context) (Principal, error) {
	current, ok := ctx.Value(principalKey{}).(Principal)
	if !ok || current.UserID == (uuid.UUID{}) || current.Token == "" {
		return Principal{}, domain.ErrUnauthorized
	}
	return current, nil
}

func CurrentLicenseToken(ctx context.Context) (string, error) {
	token, ok := ctx.Value(licenseTokenKey{}).(string)
	if !ok || token == "" {
		return "", domain.ErrUnauthorized
	}
	return token, nil
}
