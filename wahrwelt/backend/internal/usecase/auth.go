package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"uuid"
)

type Auth struct {
	users    outbound.UserRepository
	sessions outbound.SessionStore
	password outbound.PasswordHasher
	ttl      time.Duration
}

func NewAuth(users outbound.UserRepository, sessions outbound.SessionStore, password outbound.PasswordHasher, ttl time.Duration) *Auth {
	return &Auth{users: users, sessions: sessions, password: password, ttl: ttl}
}

func (a *Auth) Register(ctx context.Context, username, password string) (domain.User, error) {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return domain.User{}, err
	}
	if err := domain.ValidatePassword(password); err != nil {
		return domain.User{}, err
	}
	passwordHash, err := a.password.Hash(password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	user := domain.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}
	created, err := a.users.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return created, nil
}

func (a *Auth) Login(ctx context.Context, username, password string) (domain.Session, error) {
	username, err := domain.NormalizeUsername(username)
	if err != nil || domain.ValidateLoginPassword(password) != nil {
		return domain.Session{}, domain.ErrUnauthorized
	}
	user, err := a.users.GetUserByUsername(ctx, username)
	if errors.Is(err, domain.ErrNotFound) {
		_ = a.password.Verify("", password)
		return domain.Session{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.Session{}, fmt.Errorf("find user: %w", err)
	}
	if err := a.password.Verify(user.PasswordHash, password); err != nil {
		return domain.Session{}, domain.ErrUnauthorized
	}
	token, err := sessionToken()
	if err != nil {
		return domain.Session{}, fmt.Errorf("create session token: %w", err)
	}
	if a.ttl <= 0 {
		return domain.Session{}, fmt.Errorf("session TTL must be positive")
	}
	if err := a.sessions.CreateSession(ctx, token, user.ID, a.ttl); err != nil {
		return domain.Session{}, fmt.Errorf("store session: %w", err)
	}
	return domain.Session{Token: token, UserID: user.ID, ExpiresAt: time.Now().Add(a.ttl)}, nil
}

func (a *Auth) ResolveSession(ctx context.Context, token string) (uuid.UUID, error) {
	if token == "" {
		return uuid.UUID{}, domain.ErrUnauthorized
	}
	id, err := a.sessions.GetSession(ctx, token)
	if errors.Is(err, domain.ErrNotFound) {
		return uuid.UUID{}, domain.ErrUnauthorized
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("load session: %w", err)
	}
	if id == (uuid.UUID{}) {
		return uuid.UUID{}, domain.ErrUnauthorized
	}
	return id, nil
}

func (a *Auth) User(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	if userID == (uuid.UUID{}) {
		return domain.User{}, domain.ErrUnauthorized
	}
	user, err := a.users.GetUserByID(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("load user: %w", err)
	}
	user.PasswordHash = ""
	return user, nil
}

func (a *Auth) ListUsers(ctx context.Context) ([]domain.User, error) {
	users, err := a.users.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (a *Auth) Logout(ctx context.Context, token string) error {
	if token == "" {
		return domain.ErrUnauthorized
	}
	if err := a.sessions.DeleteSession(ctx, token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func sessionToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}
