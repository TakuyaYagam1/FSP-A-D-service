package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"uuid"
)

type Case struct{ repository outbound.CaseRepository }

func NewCase(repository outbound.CaseRepository) *Case {
	return &Case{repository: repository}
}

func (c *Case) Create(ctx context.Context, ownerID uuid.UUID, kindValue, originValue, destinationValue string) (domain.Case, error) {
	kind, ok := domain.ParseCaseKind(kindValue)
	if !ok || ownerID == (uuid.UUID{}) {
		return domain.Case{}, domain.ErrInvalidInput
	}
	origin, err := domain.ValidateRouteText(originValue)
	if err != nil {
		return domain.Case{}, err
	}
	destination, err := domain.ValidateRouteText(destinationValue)
	if err != nil {
		return domain.Case{}, err
	}
	item := domain.Case{
		ID:          uuid.New(),
		OwnerID:     ownerID,
		Kind:        kind,
		Origin:      origin,
		Destination: destination,
		CreatedAt:   time.Now().UTC(),
	}
	created, err := c.repository.CreateCase(ctx, item)
	if err != nil {
		return domain.Case{}, fmt.Errorf("create case: %w", err)
	}
	return created, nil
}

const casePageSize = 20

func (c *Case) List(ctx context.Context, actorID uuid.UUID, offset int32) ([]domain.Case, bool, error) {
	if actorID == (uuid.UUID{}) {
		return nil, false, domain.ErrNotFound
	}
	if offset < 0 {
		return nil, false, domain.ErrInvalidInput
	}
	items, err := c.repository.ListCases(ctx, actorID, offset)
	if err != nil {
		return nil, false, fmt.Errorf("list cases: %w", err)
	}
	for _, item := range items {
		if item.OwnerID != actorID {
			return nil, false, domain.ErrNotFound
		}
	}
	more := len(items) > casePageSize
	if more {
		items = items[:casePageSize]
	}
	return items, more, nil
}

func (c *Case) Get(ctx context.Context, actorID, caseID uuid.UUID) (domain.Case, error) {
	if actorID == (uuid.UUID{}) || caseID == (uuid.UUID{}) {
		return domain.Case{}, domain.ErrNotFound
	}
	item, err := c.repository.GetCase(ctx, caseID)
	if err != nil {
		return domain.Case{}, fmt.Errorf("get case: %w", err)
	}
	if item.OwnerID != actorID {
		return domain.Case{}, domain.ErrNotFound
	}
	return item, nil
}
