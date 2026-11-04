package domain

import (
	"time"

	"uuid"
)

type InspectionVisibility string

const (
	InspectionVisibilityPrivate InspectionVisibility = "private"
	InspectionVisibilityPublic  InspectionVisibility = "public"
)

type Inspection struct {
	ID             uuid.UUID
	CaseID         uuid.UUID
	OwnerID        uuid.UUID
	InspectorID    uuid.UUID
	Number         string
	InspectorNote  string
	NoteVisibility InspectionVisibility
	CreatedAt      time.Time
}

type InspectionComment struct {
	ID               uuid.UUID
	InspectionNumber string
	InspectionID     uuid.UUID
	AuthorID         uuid.UUID
	Text             string
	CreatedAt        time.Time
}

type InspectionCard struct {
	Number        string
	CaseID        uuid.UUID
	InspectorNote string
	Comments      []InspectionComment
	CreatedAt     time.Time
}
