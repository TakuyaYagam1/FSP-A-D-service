package redis

import (
	"context"
	"fmt"
	"time"

	"uuid"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"github.com/redis/go-redis/v9"
)

var recordApprovalScript = redis.NewScript(`
-- Повтор одного участника оставляет ту же подпись и не продлевает TTL
redis.call("HSETNX", KEYS[1], ARGV[1], ARGV[2])
local existing = redis.call("HGET", KEYS[1], ARGV[1])
if existing ~= ARGV[2] then
  return redis.error_reply("approval role already belongs to another user")
end
if redis.call("PTTL", KEYS[1]) < 0 then
  redis.call("PEXPIRE", KEYS[1], ARGV[3])
end
return {
  redis.call("HGET", KEYS[1], "carrier") or "",
  redis.call("HGET", KEYS[1], "inspector") or ""
}
`)

func (s *Store) RecordApproval(ctx context.Context, slotID, actorID uuid.UUID, role domain.ApprovalRole, ttl time.Duration) (domain.ApprovalState, error) {
	if slotID == (uuid.UUID{}) || actorID == (uuid.UUID{}) || !role.Valid() || ttl.Milliseconds() < 1 {
		return domain.ApprovalState{}, domain.ErrInvalidInput
	}
	// Запись роли выполняется атомарно под ключом слота в общем Redis namespace
	values, err := recordApprovalScript.Run(ctx, s.client, []string{approvalKey(slotID)}, string(role), actorID.String(), ttl.Milliseconds()).Slice()
	if err != nil {
		return domain.ApprovalState{}, fmt.Errorf("record approval: %w", err)
	}
	if len(values) != 2 {
		return domain.ApprovalState{}, fmt.Errorf("record approval: unexpected state length %d", len(values))
	}
	return approvalStateFromValues(values[0], values[1])
}

func (s *Store) GetApprovals(ctx context.Context, slotID uuid.UUID) (domain.ApprovalState, error) {
	if slotID == (uuid.UUID{}) {
		return domain.ApprovalState{}, domain.ErrInvalidInput
	}
	values, err := s.client.HGetAll(ctx, approvalKey(slotID)).Result()
	if err != nil {
		return domain.ApprovalState{}, fmt.Errorf("read approvals: %w", err)
	}
	if len(values) == 0 {
		return domain.ApprovalState{}, domain.ErrNotFound
	}
	return approvalStateFromValues(values["carrier"], values["inspector"])
}

func approvalStateFromValues(carrierValue, inspectorValue interface{}) (domain.ApprovalState, error) {
	carrierID, err := approvalUserID(carrierValue)
	if err != nil {
		return domain.ApprovalState{}, fmt.Errorf("parse carrier approval: %w", err)
	}
	inspectorID, err := approvalUserID(inspectorValue)
	if err != nil {
		return domain.ApprovalState{}, fmt.Errorf("parse inspector approval: %w", err)
	}
	return domain.ApprovalState{CarrierID: carrierID, InspectorID: inspectorID}, nil
}

func approvalUserID(value interface{}) (uuid.UUID, error) {
	if value == nil {
		return uuid.UUID{}, nil
	}
	var text string
	switch raw := value.(type) {
	case string:
		text = raw
	case []byte:
		text = string(raw)
	default:
		return uuid.UUID{}, fmt.Errorf("unexpected value type %T", value)
	}
	if text == "" {
		return uuid.UUID{}, nil
	}
	id, err := uuid.Parse(text)
	if err != nil || id == (uuid.UUID{}) {
		return uuid.UUID{}, fmt.Errorf("invalid user id")
	}
	return id, nil
}

func approvalKey(slotID uuid.UUID) string {
	return namespace + ":approval-records:" + slotID.String()
}

var _ outbound.ApprovalStore = (*Store)(nil)
