package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Order is the aggregate root for the order domain.
type Order struct {
	ID           string
	TenantID     string
	CustomerName string
	Status       string
	CreatedAt    time.Time
}

// Allowed statuses.
const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"
)

// NewOrder creates a new order.
func NewOrder(tenantID, customerName string) (*Order, error) {
	if tenantID == "" {
		return nil, errors.New("tenant id is required")
	}
	if customerName == "" {
		return nil, errors.New("customer name is required")
	}
	return &Order{
		ID:           uuid.NewString(),
		TenantID:     tenantID,
		CustomerName: customerName,
		Status:       StatusPending,
		CreatedAt:    time.Now().UTC(),
	}, nil
}

// Confirm transitions an order to confirmed status.
func (o *Order) Confirm() error {
	if o.Status != StatusPending {
		return errors.New("only pending orders can be confirmed")
	}
	o.Status = StatusConfirmed
	return nil
}
