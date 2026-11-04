package postgres

import (
	"context"
	"fmt"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

func (s *Store) CreateInspection(ctx context.Context, item domain.Inspection) (domain.Inspection, error) {
	row, err := s.queries.CreateInspection(ctx, sqlc.CreateInspectionParams{
		ID:               item.ID,
		CaseID:           item.CaseID,
		OwnerID:          item.OwnerID,
		InspectorID:      item.InspectorID,
		InspectionNumber: item.Number,
		InspectorNote:    item.InspectorNote,
		NoteVisibility:   string(item.NoteVisibility),
	})
	if err != nil {
		return domain.Inspection{}, translate(err)
	}
	return fromInspection(row), nil
}

func (s *Store) GetInspectionByNumber(ctx context.Context, number string) (domain.Inspection, error) {
	row, err := s.queries.GetInspectionByNumber(ctx, number)
	if err != nil {
		return domain.Inspection{}, translate(err)
	}
	return fromInspection(row), nil
}

func (s *Store) ListInspectionComments(ctx context.Context, inspectionID uuid.UUID) ([]domain.InspectionComment, error) {
	rows, err := s.queries.ListInspectionComments(ctx, inspectionID)
	if err != nil {
		return nil, translate(err)
	}
	result := make([]domain.InspectionComment, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InspectionComment{
			ID:           row.ID,
			InspectionID: row.InspectionID,
			AuthorID:     row.AuthorID,
			Text:         row.Text,
			CreatedAt:    row.CreatedAt.Time,
		})
	}
	return result, nil
}

func (s *Store) AddInspectionComment(ctx context.Context, comment domain.InspectionComment) (domain.InspectionComment, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.InspectionComment{}, fmt.Errorf("begin inspection comment transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := s.queries.WithTx(tx)
	inspection, err := queries.GetInspectionByNumber(ctx, comment.InspectionNumber)
	if err != nil {
		return domain.InspectionComment{}, translate(err)
	}
	comment.InspectionID = inspection.ID
	row, err := queries.CreateInspectionComment(ctx, sqlc.CreateInspectionCommentParams{
		ID:           comment.ID,
		InspectionID: comment.InspectionID,
		AuthorID:     comment.AuthorID,
		Text:         comment.Text,
	})
	if err != nil {
		return domain.InspectionComment{}, translate(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.InspectionComment{}, fmt.Errorf("commit inspection comment transaction: %w", err)
	}
	return domain.InspectionComment{
		ID:           row.ID,
		InspectionID: row.InspectionID,
		AuthorID:     row.AuthorID,
		Text:         row.Text,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

func fromInspection(row sqlc.Inspection) domain.Inspection {
	return domain.Inspection{
		ID:             row.ID,
		CaseID:         row.CaseID,
		OwnerID:        row.OwnerID,
		InspectorID:    row.InspectorID,
		Number:         row.InspectionNumber,
		InspectorNote:  row.InspectorNote,
		NoteVisibility: domain.InspectionVisibility(row.NoteVisibility),
		CreatedAt:      row.CreatedAt.Time,
	}
}
