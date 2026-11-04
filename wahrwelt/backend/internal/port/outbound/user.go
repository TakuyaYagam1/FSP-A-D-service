package outbound

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

type UserRepository interface {
	CreateUser(context.Context, domain.User) (domain.User, error)
	GetUserByID(context.Context, uuid.UUID) (domain.User, error)
	GetUserByUsername(context.Context, string) (domain.User, error)
	ListUsers(context.Context) ([]domain.User, error)
}

type PasswordHasher interface {
	Hash(string) (string, error)
	Verify(string, string) error
}
