package domain

import (
	"time"

	"uuid"
)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type Session struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
}
