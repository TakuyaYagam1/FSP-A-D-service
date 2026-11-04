package v1

import (
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"uuid"
)

func parseID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, domain.ErrInvalidInput
	}
	return id, nil
}
