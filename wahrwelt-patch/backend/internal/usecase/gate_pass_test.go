package usecase

import (
	"context"
	"errors"
	"testing"

	"uuid"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

type gatePassRepositoryStub struct {
	item domain.GatePass
}

func (r *gatePassRepositoryStub) CreateGatePass(_ context.Context, item domain.GatePass) (domain.GatePass, error) {
	r.item = item
	return item, nil
}

func (r *gatePassRepositoryStub) GetGatePassByNumber(_ context.Context, _ string) (domain.GatePass, error) {
	return r.item, nil
}

func TestGetSealedUsesSessionIdentity(t *testing.T) {
	ownerID := uuid.New()
	operatorID := uuid.New()
	outsiderID := uuid.New()
	pass := domain.GatePass{
		OwnerID:        ownerID,
		GateOperatorID: operatorID,
		Number:         "GATE-test",
		GateWord:       "sealed value",
	}
	service := NewGatePass(&gatePassRepositoryStub{item: pass}, nil)

	tests := []struct {
		name     string
		userID   uuid.UUID
		wantWord string
		wantErr  error
	}{
		{name: "owner", userID: ownerID, wantWord: pass.GateWord},
		{name: "assigned operator", userID: operatorID, wantWord: pass.GateWord},
		{name: "unassigned user", userID: outsiderID, wantErr: domain.ErrNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := service.GetSealed(context.Background(), test.userID, pass.Number)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("GetSealed() error = %v, want %v", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetSealed() error = %v", err)
			}
			if got.GateWord != test.wantWord {
				t.Fatalf("GetSealed() gate word = %q, want %q", got.GateWord, test.wantWord)
			}
		})
	}
}
