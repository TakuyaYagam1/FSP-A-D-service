package usecase

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"context"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"uuid"
)

const LicenseReadScope = "license:read"

type License struct {
	repository outbound.LicenseRepository
	secret     []byte
	tokenTTL   time.Duration
}

type licenseTokenPayload struct {
	ExpiresAt int64  `json:"exp"`
	Scope     string `json:"scope"`
	Nonce     string `json:"nonce"`
}

func NewLicense(repository outbound.LicenseRepository, secret string, tokenTTL time.Duration) *License {
	return &License{repository: repository, secret: []byte(secret), tokenTTL: tokenTTL}
}

func (l *License) Create(ctx context.Context, ownerID uuid.UUID, pin string) (domain.License, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.License{}, domain.ErrUnauthorized
	}
	pin = strings.TrimSpace(pin)
	if pin == "" || len(pin) > 240 {
		return domain.License{}, domain.ErrInvalidInput
	}
	id := uuid.New()
	item := domain.License{
		ID:         id,
		OwnerID:    ownerID,
		Number:     "LIC-" + id.String(),
		Status:     domain.LicenseStatusActive,
		ReleasePin: pin,
		CreatedAt:  time.Now().UTC(),
	}
	created, err := l.repository.CreateLicense(ctx, item)
	if err != nil {
		return domain.License{}, fmt.Errorf("create export license: %w", err)
	}
	return created, nil
}

func (l *License) PublicCard(ctx context.Context, number string) (domain.LicenseCard, error) {
	item, err := l.repository.GetLicenseByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.LicenseCard{}, fmt.Errorf("get export license card: %w", err)
	}
	return domain.LicenseCard{Number: item.Number, Status: item.Status}, nil
}

func (l *License) Get(ctx context.Context, ownerID uuid.UUID, number string) (domain.License, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.License{}, domain.ErrUnauthorized
	}
	item, err := l.repository.GetLicenseByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.License{}, fmt.Errorf("get export license: %w", err)
	}
	if item.OwnerID != ownerID {
		return domain.License{}, domain.ErrNotFound
	}
	return item, nil
}

func (l *License) IssueToken(ctx context.Context, ownerID uuid.UUID, number, scope string) (domain.LicenseToken, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.LicenseToken{}, domain.ErrUnauthorized
	}
	if scope != LicenseReadScope {
		return domain.LicenseToken{}, domain.ErrInvalidInput
	}
	item, err := l.repository.GetLicenseByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.LicenseToken{}, fmt.Errorf("get export license for token: %w", err)
	}
	if item.OwnerID != ownerID {
		return domain.LicenseToken{}, domain.ErrNotFound
	}
	return l.signToken(scope, time.Now().UTC().Add(l.tokenTTL).Truncate(time.Second))
}

func (l *License) ReadWithToken(ctx context.Context, token, number string) (domain.License, error) {
	if err := l.ValidateReadToken(token); err != nil {
		return domain.License{}, err
	}
	item, err := l.repository.GetLicenseByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.License{}, fmt.Errorf("get export license with access token: %w", err)
	}
	return item, nil
}

func (l *License) ValidateReadToken(token string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return domain.ErrUnauthorized
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return domain.ErrUnauthorized
	}
	mac := hmac.New(sha256.New, l.secret)
	_, _ = mac.Write([]byte(parts[0]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return domain.ErrUnauthorized
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return domain.ErrUnauthorized
	}
	var claims map[string]json.RawMessage
	if err := json.Unmarshal(payload, &claims); err != nil || len(claims) != 3 || claims["exp"] == nil || claims["scope"] == nil || claims["nonce"] == nil {
		return domain.ErrUnauthorized
	}
	var decoded licenseTokenPayload
	if err := json.Unmarshal(payload, &decoded); err != nil || decoded.Scope != LicenseReadScope || decoded.ExpiresAt <= time.Now().Unix() {
		return domain.ErrUnauthorized
	}
	nonce, err := base64.RawURLEncoding.DecodeString(decoded.Nonce)
	if err != nil || len(nonce) != 16 {
		return domain.ErrUnauthorized
	}
	return nil
}

func (l *License) signToken(scope string, expiresAt time.Time) (domain.LicenseToken, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return domain.LicenseToken{}, fmt.Errorf("create license token nonce: %w", err)
	}
	claims := licenseTokenPayload{
		ExpiresAt: expiresAt.Unix(),
		Scope:     scope,
		Nonce:     base64.RawURLEncoding.EncodeToString(nonce),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return domain.LicenseToken{}, fmt.Errorf("encode license token: %w", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, l.secret)
	_, _ = mac.Write([]byte(encoded))
	token := encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return domain.LicenseToken{Value: token, ExpiresAt: expiresAt}, nil
}
