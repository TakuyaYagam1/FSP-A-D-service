package api

import (
	"net/http/httptest"
	"testing"
)

func TestRoutePassCanonicalSlotID(t *testing.T) {
	const slotID = "123e4567-e89b-12d3-a456-426614174000"
	request := httptest.NewRequest("POST", "/api/v1/route-slots/"+slotID+"/pass", nil)
	params, err := decodeIssueRoutePassParams([1]string{slotID}, false, request)
	if err != nil {
		t.Fatalf("canonical route slot UUID rejected by issue endpoint: %v", err)
	}
	if params.SlotID != slotID {
		t.Fatalf("decoded issue slot ID = %q, want %q", params.SlotID, slotID)
	}
	getRequest := httptest.NewRequest("GET", "/api/v1/route-slots/"+slotID+"/pass", nil)
	getParams, err := decodeGetRoutePassParams([1]string{slotID}, false, getRequest)
	if err != nil {
		t.Fatalf("canonical route slot UUID rejected by get endpoint: %v", err)
	}
	if getParams.SlotID != slotID {
		t.Fatalf("decoded get slot ID = %q, want %q", getParams.SlotID, slotID)
	}
}
