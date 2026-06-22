package domain

import "context"

// Repository defines persistence operations for Order.
type Repository interface {
	Save(ctx context.Context, order *Order) error
	Get(ctx context.Context, id string) (*Order, error)
	ListByStatus(ctx context.Context, tenantID, status string) ([]*Order, error)
}
