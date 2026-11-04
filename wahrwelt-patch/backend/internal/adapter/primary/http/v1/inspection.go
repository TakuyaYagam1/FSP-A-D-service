package v1

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
)

func (h *Handler) CreateInspection(ctx context.Context, req *api.CreateInspectionRequest) (*api.InspectionCreated, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	caseID, err := parseID(req.CaseID)
	if err != nil {
		return nil, err
	}
	item, err := h.features.Inspection.Create(ctx, current.UserID, caseID, req.InspectorUsername, req.InspectorNote)
	if err != nil {
		return nil, err
	}
	return &api.InspectionCreated{InspectionNumber: item.Number, CreatedAt: item.CreatedAt}, nil
}

func (h *Handler) GetInspection(ctx context.Context, params api.GetInspectionParams) (*api.InspectionCard, error) {
	item, err := h.features.Inspection.PublicCard(ctx, params.InspectionNumber)
	if err != nil {
		return nil, err
	}
	comments := make([]api.InspectionComment, 0, len(item.Comments))
	for _, comment := range item.Comments {
		comments = append(comments, api.InspectionComment{
			CommentID: comment.ID.String(),
			Text:      comment.Text,
			CreatedAt: comment.CreatedAt,
		})
	}
	response := &api.InspectionCard{
		InspectionNumber: item.Number,
		CaseID:           item.CaseID.String(),
		Comments:         comments,
		CreatedAt:        item.CreatedAt,
	}
	if item.InspectorNote != "" {
		response.InspectorNote = api.NewOptString(item.InspectorNote)
	}
	return response, nil
}

func (h *Handler) GetInspectionNote(ctx context.Context, params api.GetInspectionNoteParams) (*api.InspectionNote, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	note, err := h.features.Inspection.PrivateNote(ctx, current.UserID, params.InspectionNumber)
	if err != nil {
		return nil, err
	}
	return &api.InspectionNote{InspectorNote: note}, nil
}

func (h *Handler) AddInspectionComment(ctx context.Context, req *api.AddInspectionCommentRequest, params api.AddInspectionCommentParams) (*api.InspectionComment, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	// Публичный комментарий не должен менять доступ к закрытой заметке акта
	comment, err := h.features.Inspection.AddComment(ctx, current.UserID, params.InspectionNumber, req.Text)
	if err != nil {
		return nil, err
	}
	return &api.InspectionComment{
		CommentID: comment.ID.String(),
		Text:      comment.Text,
		CreatedAt: comment.CreatedAt,
	}, nil
}
