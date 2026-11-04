package v1

import (
	"context"
	"testing"
	"time"

	"uuid"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"github.com/TakuyaYagam1/wahrwelt/internal/usecase"
)

type inspectionSessionStoreStub struct{ userID uuid.UUID }

func (s *inspectionSessionStoreStub) CreateSession(context.Context, string, uuid.UUID, time.Duration) error {
	return nil
}

func (s *inspectionSessionStoreStub) GetSession(context.Context, string) (uuid.UUID, error) {
	return s.userID, nil
}

func (s *inspectionSessionStoreStub) DeleteSession(context.Context, string) error { return nil }

type inspectionRepositoryStub struct {
	inspection domain.Inspection
}

func (r *inspectionRepositoryStub) CreateInspection(_ context.Context, item domain.Inspection) (domain.Inspection, error) {
	return item, nil
}

func (r *inspectionRepositoryStub) GetInspectionByNumber(context.Context, string) (domain.Inspection, error) {
	return r.inspection, nil
}

func (r *inspectionRepositoryStub) ListInspectionComments(context.Context, uuid.UUID) ([]domain.InspectionComment, error) {
	return nil, nil
}

func (r *inspectionRepositoryStub) AddInspectionComment(_ context.Context, comment domain.InspectionComment) (domain.InspectionComment, error) {
	return comment, nil
}

var _ outbound.SessionStore = (*inspectionSessionStoreStub)(nil)
var _ outbound.InspectionRepository = (*inspectionRepositoryStub)(nil)

func TestAddInspectionCommentDoesNotChangeVisibility(t *testing.T) {
	userID := uuid.New()
	inspectionID := uuid.New()
	repository := &inspectionRepositoryStub{inspection: domain.Inspection{
		ID:             inspectionID,
		OwnerID:        userID,
		InspectorID:    uuid.New(),
		Number:         "INSP-" + inspectionID.String(),
		InspectorNote:  "private note",
		NoteVisibility: domain.InspectionVisibilityPrivate,
	}}
	sessions := &inspectionSessionStoreStub{userID: userID}
	auth := usecase.NewAuth(nil, sessions, nil, time.Minute)
	security := middleware.NewSecurityHandler(auth, nil)
	ctx, err := security.HandleSessionCookie(context.Background(), api.AddInspectionCommentOperation, api.SessionCookie{APIKey: "session"})
	if err != nil {
		t.Fatalf("HandleSessionCookie() error = %v", err)
	}
	handler := NewHandler(nil, nil, nil, FeatureSet{Inspection: usecase.NewInspection(repository, nil, nil)}, time.Minute, false)
	request := &api.AddInspectionCommentRequest{
		Text:       "cargo checked",
		Visibility: api.NewOptAddInspectionCommentRequestVisibility(api.AddInspectionCommentRequestVisibilityPublic),
	}
	params := api.AddInspectionCommentParams{InspectionNumber: repository.inspection.Number}

	if _, err := handler.AddInspectionComment(ctx, request, params); err != nil {
		t.Fatalf("AddInspectionComment() error = %v", err)
	}
	if repository.inspection.NoteVisibility != domain.InspectionVisibilityPrivate {
		t.Fatalf("inspection visibility = %q, want %q", repository.inspection.NoteVisibility, domain.InspectionVisibilityPrivate)
	}
}
