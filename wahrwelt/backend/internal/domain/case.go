package domain

import (
	"time"

	"uuid"
)

type CaseKind string

const (
	CaseKindImport  CaseKind = "import"
	CaseKindExport  CaseKind = "export"
	CaseKindService CaseKind = "service"
)

func ParseCaseKind(value string) (CaseKind, bool) {
	kind := CaseKind(value)
	switch kind {
	case CaseKindImport, CaseKindExport, CaseKindService:
		return kind, true
	default:
		return "", false
	}
}

type Case struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Kind        CaseKind
	Origin      string
	Destination string
	CreatedAt   time.Time
}
