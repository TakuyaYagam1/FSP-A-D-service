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

type Inspection struct {
	repository outbound.InspectionRepository
	cases      outbound.CaseRepository
	users      outbound.UserRepository
}

func NewInspection(repository outbound.InspectionRepository, cases outbound.CaseRepository, users outbound.UserRepository) *Inspection {
	return &Inspection{repository: repository, cases: cases, users: users}
}

func (i *Inspection) Create(ctx context.Context, ownerID, caseID uuid.UUID, inspectorUsername, note string) (domain.Inspection, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.Inspection{}, domain.ErrUnauthorized
	}
	caseItem, err := i.cases.GetCase(ctx, caseID)
	if err != nil {
		return domain.Inspection{}, fmt.Errorf("get inspection case: %w", err)
	}
	if caseItem.OwnerID != ownerID {
		return domain.Inspection{}, domain.ErrNotFound
	}
	inspector, err := i.users.GetUserByUsername(ctx, strings.TrimSpace(inspectorUsername))
	if err != nil {
		return domain.Inspection{}, fmt.Errorf("get assigned inspector: %w", err)
	}
	if inspector.ID == ownerID {
		return domain.Inspection{}, domain.ErrInvalidInput
	}
	note = strings.TrimSpace(note)
	if note == "" || len(note) > 500 {
		return domain.Inspection{}, domain.ErrInvalidInput
	}
	id := uuid.New()
	item := domain.Inspection{
		ID:             id,
		CaseID:         caseItem.ID,
		OwnerID:        ownerID,
		InspectorID:    inspector.ID,
		Number:         "INSP-" + id.String(),
		InspectorNote:  note,
		NoteVisibility: domain.InspectionVisibilityPrivate,
		CreatedAt:      time.Now().UTC(),
	}
	created, err := i.repository.CreateInspection(ctx, item)
	if err != nil {
		return domain.Inspection{}, fmt.Errorf("create inspection: %w", err)
	}
	return created, nil
}

func (i *Inspection) PublicCard(ctx context.Context, number string) (domain.InspectionCard, error) {
	item, err := i.repository.GetInspectionByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.InspectionCard{}, fmt.Errorf("get inspection: %w", err)
	}
	comments, err := i.repository.ListInspectionComments(ctx, item.ID)
	if err != nil {
		return domain.InspectionCard{}, fmt.Errorf("list inspection comments: %w", err)
	}
	card := domain.InspectionCard{Number: item.Number, CaseID: item.CaseID, Comments: comments, CreatedAt: item.CreatedAt}
	if item.NoteVisibility == domain.InspectionVisibilityPublic {
		card.InspectorNote = item.InspectorNote
	}
	return card, nil
}

func (i *Inspection) PrivateNote(ctx context.Context, actorID uuid.UUID, number string) (string, error) {
	if actorID == (uuid.UUID{}) {
		return "", domain.ErrUnauthorized
	}
	item, err := i.repository.GetInspectionByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return "", fmt.Errorf("get inspection note: %w", err)
	}
	if item.OwnerID != actorID && item.InspectorID != actorID {
		return "", domain.ErrNotFound
	}
	return item.InspectorNote, nil
}

func (i *Inspection) AddComment(ctx context.Context, actorID uuid.UUID, number, value string) (domain.InspectionComment, error) {
	if actorID == (uuid.UUID{}) {
		return domain.InspectionComment{}, domain.ErrUnauthorized
	}
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 500 {
		return domain.InspectionComment{}, domain.ErrInvalidInput
	}
	comment := domain.InspectionComment{
		ID:               uuid.New(),
		InspectionNumber: strings.TrimSpace(number),
		AuthorID:         actorID,
		Text:             value,
		CreatedAt:        time.Now().UTC(),
	}
	created, err := i.repository.AddInspectionComment(ctx, comment)
	if err != nil {
		return domain.InspectionComment{}, fmt.Errorf("add inspection comment: %w", err)
	}
	return created, nil
}
