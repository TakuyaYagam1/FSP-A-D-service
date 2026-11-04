package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"uuid"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
)

type GatePass struct {
	repository outbound.GatePassRepository
	users      outbound.UserRepository
}

func NewGatePass(repository outbound.GatePassRepository, users outbound.UserRepository) *GatePass {
	return &GatePass{repository: repository, users: users}
}

func (g *GatePass) Create(ctx context.Context, ownerID uuid.UUID, operatorUsername, state, gateWord string) (domain.GatePass, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.GatePass{}, domain.ErrUnauthorized
	}
	operatorUsername = strings.TrimSpace(operatorUsername)
	gateWord = strings.TrimSpace(gateWord)
	if len(operatorUsername) < 3 || len(operatorUsername) > 32 || gateWord == "" || len(gateWord) > 240 {
		return domain.GatePass{}, domain.ErrInvalidInput
	}
	if state != "issued" && state != "closed" {
		return domain.GatePass{}, domain.ErrInvalidInput
	}
	operator, err := g.users.GetUserByUsername(ctx, operatorUsername)
	if err != nil {
		return domain.GatePass{}, fmt.Errorf("get gate operator: %w", err)
	}
	if operator.ID == ownerID {
		return domain.GatePass{}, domain.ErrInvalidInput
	}
	id := uuid.New()
	item := domain.GatePass{
		ID:             id,
		OwnerID:        ownerID,
		GateOperatorID: operator.ID,
		Number:         "GATE-" + id.String(),
		State:          state,
		GateWord:       gateWord,
		CreatedAt:      time.Now().UTC(),
	}
	created, err := g.repository.CreateGatePass(ctx, item)
	if err != nil {
		return domain.GatePass{}, fmt.Errorf("create gate pass: %w", err)
	}
	return created, nil
}

func (g *GatePass) Get(ctx context.Context, userID uuid.UUID, number string) (domain.GatePass, error) {
	if userID == (uuid.UUID{}) {
		return domain.GatePass{}, domain.ErrUnauthorized
	}
	item, err := g.repository.GetGatePassByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.GatePass{}, fmt.Errorf("get gate pass: %w", err)
	}
	if item.OwnerID != userID && item.GateOperatorID != userID {
		return domain.GatePass{}, domain.ErrNotFound
	}
	return item, nil
}

func (g *GatePass) PublicCard(ctx context.Context, number string) (domain.GatePassCard, error) {
	item, err := g.repository.GetGatePassByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.GatePassCard{}, fmt.Errorf("get gate pass card: %w", err)
	}
	return domain.GatePassCard{Number: item.Number, State: item.State}, nil
}

func (g *GatePass) GetSealed(ctx context.Context, userID uuid.UUID, number string) (domain.GatePass, error) {
	// Get сверяет сессию с владельцем и назначенным оператором этого пропуска
	return g.Get(ctx, userID, number)
}
