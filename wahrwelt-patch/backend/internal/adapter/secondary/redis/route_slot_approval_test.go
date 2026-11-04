package redis

import (
	"context"
	"os"
	"testing"
	"time"

	"uuid"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/redis/go-redis/v9"
)

func TestRouteSlotApprovalsAreRoleBoundAndIdempotent(t *testing.T) {
	address := os.Getenv("TEST_REDIS_ADDR")
	if address == "" {
		t.Skip("set TEST_REDIS_ADDR to run the Redis approval integration test")
	}

	client := redis.NewClient(&redis.Options{Addr: address})
	t.Cleanup(func() { _ = client.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("connect to test Redis: %v", err)
	}

	store := &Store{client: client}
	slotID := uuid.New()
	carrierID := uuid.New()
	inspectorID := uuid.New()
	state, err := store.RecordApproval(ctx, slotID, carrierID, domain.RoleCarrier, 10*time.Second)
	if err != nil {
		t.Fatalf("record carrier approval: %v", err)
	}
	if state.CarrierID != carrierID || state.InspectorID != (uuid.UUID{}) || state.Complete() {
		t.Fatalf("first approval state = %+v", state)
	}

	state, err = store.RecordApproval(ctx, slotID, carrierID, domain.RoleCarrier, time.Hour)
	if err != nil {
		t.Fatalf("repeat carrier approval: %v", err)
	}
	if state.CarrierID != carrierID || state.InspectorID != (uuid.UUID{}) || state.Complete() {
		t.Fatalf("repeated approval changed state: %+v", state)
	}
	ttl, err := client.PTTL(ctx, approvalKey(slotID)).Result()
	if err != nil {
		t.Fatalf("read approval TTL: %v", err)
	}
	if ttl <= 0 || ttl > time.Minute {
		t.Fatalf("repeated approval extended TTL: %s", ttl)
	}
	if _, err := store.RecordApproval(ctx, slotID, uuid.New(), domain.RoleCarrier, time.Hour); err == nil {
		t.Fatal("another user replaced the carrier approval")
	}

	state, err = store.RecordApproval(ctx, slotID, inspectorID, domain.RoleInspector, time.Hour)
	if err != nil {
		t.Fatalf("record inspector approval: %v", err)
	}
	if !state.Complete() || state.CarrierID != carrierID || state.InspectorID != inspectorID {
		t.Fatalf("both role approvals state = %+v", state)
	}
	state, err = store.GetApprovals(ctx, slotID)
	if err != nil {
		t.Fatalf("read approvals: %v", err)
	}
	if !state.Complete() || state.CarrierID != carrierID || state.InspectorID != inspectorID {
		t.Fatalf("stored approval state = %+v", state)
	}
}
