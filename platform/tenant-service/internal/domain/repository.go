package domain

import "context"

// Repository defines persistence operations for tenants.
type Repository interface {
	Create(ctx context.Context, name string) (*Tenant, error)
	List(ctx context.Context) ([]*Tenant, error)
}
