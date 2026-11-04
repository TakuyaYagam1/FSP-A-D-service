package domain

import (
	"time"

	"uuid"
)

type InvoiceStatus string

const InvoiceStatusPending InvoiceStatus = "pending"

type Invoice struct {
	ID         uuid.UUID
	OwnerID    uuid.UUID
	Number     string
	Amount     int64
	Status     InvoiceStatus
	PaymentRef string
	CreatedAt  time.Time
}

type InvoiceSummary struct {
	Number    string
	Amount    int64
	Status    InvoiceStatus
	CreatedAt time.Time
}

type InvoiceSearchItem struct {
	Number string
	Amount int64
	Status string
}
