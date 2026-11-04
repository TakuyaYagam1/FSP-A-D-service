package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"github.com/redis/go-redis/v9"
	"uuid"
)

var incrementApprovalScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return count
`)

func (s *Store) Increment(ctx context.Context, slotID uuid.UUID, ttl time.Duration) (domain.ApprovalState, error) {
	if slotID == (uuid.UUID{}) || ttl.Milliseconds() < 1 {
		return domain.ApprovalState{}, domain.ErrInvalidInput
	}
	count, err := incrementApprovalScript.Run(ctx, s.client, []string{approvalKey(slotID)}, ttl.Milliseconds()).Int64()
	if err != nil {
		return domain.ApprovalState{}, fmt.Errorf("record approval: %w", err)
	}
	return domain.ApprovalState{Count: count}, nil
}

func (s *Store) GetApprovals(ctx context.Context, slotID uuid.UUID) (domain.ApprovalState, error) {
	if slotID == (uuid.UUID{}) {
		return domain.ApprovalState{}, domain.ErrInvalidInput
	}
	value, err := s.client.Get(ctx, approvalKey(slotID)).Result()
	if errors.Is(err, redis.Nil) {
		return domain.ApprovalState{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ApprovalState{}, fmt.Errorf("read approvals: %w", err)
	}
	count, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return domain.ApprovalState{}, fmt.Errorf("parse approval count: %w", err)
	}
	return domain.ApprovalState{Count: count}, nil
}

func approvalKey(slotID uuid.UUID) string {
	return namespace + ":approval-count:" + slotID.String()
}

var _ outbound.ApprovalStore = (*Store)(nil)
