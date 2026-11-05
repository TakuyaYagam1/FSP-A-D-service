package postgres

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

func (s *Store) CreateRouteSlot(ctx context.Context, item domain.RouteSlot) (domain.RouteSlot, error) {
	row, err := s.queries.CreateRouteSlot(ctx, sqlc.CreateRouteSlotParams{
		ID:          item.ID,
		CaseID:      item.CaseID,
		OwnerID:     item.OwnerID,
		CarrierID:   item.CarrierID,
		InspectorID: item.InspectorID,
		SlotCode:    item.Code,
	})
	if err != nil {
		return domain.RouteSlot{}, translate(err)
	}
	return fromRouteSlot(row), nil
}

func (s *Store) GetRouteSlot(ctx context.Context, id uuid.UUID) (domain.RouteSlot, error) {
	row, err := s.queries.GetRouteSlot(ctx, id)
	if err != nil {
		return domain.RouteSlot{}, translate(err)
	}
	return fromRouteSlot(row), nil
}

func (s *Store) ListRouteSlots(ctx context.Context, caseID, ownerID uuid.UUID) ([]domain.RouteSlot, error) {
	rows, err := s.queries.ListRouteSlotsByCase(ctx, sqlc.ListRouteSlotsByCaseParams{CaseID: caseID, OwnerID: ownerID})
	if err != nil {
		return nil, translate(err)
	}
	items := make([]domain.RouteSlot, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.RouteSlot{
			ID: row.ID, CaseID: row.CaseID, OwnerID: row.OwnerID,
			CarrierID: row.CarrierID, InspectorID: row.InspectorID,
			CarrierUsername: row.CarrierUsername, InspectorUsername: row.InspectorUsername,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return items, nil
}

func fromRouteSlot(row sqlc.RouteSlot) domain.RouteSlot {
	return domain.RouteSlot{
		ID:          row.ID,
		CaseID:      row.CaseID,
		OwnerID:     row.OwnerID,
		CarrierID:   row.CarrierID,
		InspectorID: row.InspectorID,
		Code:        row.SlotCode,
		CreatedAt:   row.CreatedAt.Time,
	}
}
