package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

func TestRoutePassEndpointsRequireSession(t *testing.T) {
	server, err := NewServer(&Handler{}, middleware.NewSecurityHandler(nil, nil))
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			request := httptest.NewRequest(method, "/api/v1/route-slots/123e4567-e89b-12d3-a456-426614174000/pass", nil)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("unauthed %s status=%d body=%s", method, response.Code, response.Body.String())
			}
			if got := response.Header().Get("Cache-Control"); got != "no-store" {
				t.Fatalf("route pass %s cache policy=%q, want no-store", method, got)
			}
		})
	}
}

func TestRoutePassResponseExcludesSlotCode(t *testing.T) {
	pass := domain.RoutePass{
		SlotID:            uuid.New(),
		CaseID:            uuid.New(),
		Kind:              "import",
		Origin:            "Орел",
		Destination:       "Тула",
		CarrierUsername:   "carrier",
		InspectorUsername: "inspector",
		IssuedAt:          time.Date(2026, time.November, 5, 17, 0, 0, 0, time.UTC),
	}
	response := apiRoutePassState(domain.RoutePassState{Pass: &pass})
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("encode route pass: %v", err)
	}
	payload := string(encoded)
	for _, field := range []string{pass.SlotID.String(), pass.CaseID.String(), "carrier", "inspector", "2026-11-05T17:00:00Z"} {
		if !strings.Contains(payload, field) {
			t.Fatalf("route pass response omitted %q: %s", field, payload)
		}
	}
	if strings.Contains(payload, "slot_code") || strings.Contains(payload, "private-code") {
		t.Fatalf("route pass response leaked slot code: %s", payload)
	}
}
