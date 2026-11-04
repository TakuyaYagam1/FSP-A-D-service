package outbound

import (
	"context"
	"time"

	"uuid"
)

type SessionStore interface {
	CreateSession(context.Context, string, uuid.UUID, time.Duration) error
	GetSession(context.Context, string) (uuid.UUID, error)
	DeleteSession(context.Context, string) error
}
