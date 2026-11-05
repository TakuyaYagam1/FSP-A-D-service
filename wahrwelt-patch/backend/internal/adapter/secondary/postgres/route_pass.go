package postgres

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

func (s *Store) GetRoutePass(ctx context.Context, slotID uuid.UUID) (domain.RoutePass, error) {
	row, err := s.queries.GetRoutePass(ctx, slotID)
	if err != nil {
		return domain.RoutePass{}, translate(err)
	}
	return domain.RoutePass{
		SlotID:            row.RouteSlotID,
		CaseID:            row.CaseID,
		Kind:              row.Kind,
		Origin:            row.Origin,
		Destination:       row.Destination,
		CarrierUsername:   row.CarrierUsername,
		InspectorUsername: row.InspectorUsername,
		IssuedAt:          row.IssuedAt.Time,
	}, nil
}

func (s *Store) CreateRoutePass(ctx context.Context, slotID uuid.UUID) (domain.RoutePass, error) {
	if _, err := s.queries.CreateRoutePass(ctx, slotID); err != nil {
		return domain.RoutePass{}, translate(err)
	}
	return s.GetRoutePass(ctx, slotID)
}
