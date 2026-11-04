package postgres

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

func (s *Store) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	row, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		return domain.User{}, translate(err)
	}
	return domain.User{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return domain.User{}, translate(err)
	}
	return domain.User{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	row, err := s.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return domain.User{}, translate(err)
	}
	return domain.User{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

func (s *Store) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := s.queries.ListUsers(ctx)
	if err != nil {
		return nil, translate(err)
	}
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, domain.User{
			ID:        row.ID,
			Username:  row.Username,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return users, nil
}
