package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"uuid"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"github.com/redis/go-redis/v9"
)

func (s *Store) CreateSession(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error {
	if token == "" || userID == (uuid.UUID{}) || ttl <= 0 {
		return domain.ErrInvalidInput
	}
	if err := s.client.Set(ctx, sessionKey(token), userID.String(), ttl).Err(); err != nil {
		return fmt.Errorf("store session: %w", err)
	}
	return nil
}

func (s *Store) GetSession(ctx context.Context, token string) (uuid.UUID, error) {
	if token == "" {
		return uuid.UUID{}, domain.ErrNotFound
	}
	value, err := s.client.Get(ctx, sessionKey(token)).Result()
	if errors.Is(err, redis.Nil) {
		return uuid.UUID{}, domain.ErrNotFound
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("read session: %w", err)
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse session owner: %w", err)
	}
	return id, nil
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.client.Del(ctx, sessionKey(token)).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func sessionKey(token string) string {
	return namespace + ":session:" + token
}

var _ outbound.SessionStore = (*Store)(nil)
