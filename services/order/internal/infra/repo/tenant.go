package repo

import (
	"context"

	kitctx "github.com/plantx/kit/kit-go/context"
	"github.com/plantx/services/order/internal/domain"
)

// TenantRepo wraps a domain.Repository and enforces tenant isolation.
type TenantRepo struct {
	inner domain.Repository
}

// NewTenantRepo creates a tenant-isolating repository wrapper.
func NewTenantRepo(inner domain.Repository) *TenantRepo {
	return &TenantRepo{inner: inner}
}

func (r *TenantRepo) tenant(ctx context.Context) string {
	return kitctx.GetTenant(ctx).ID
}

// Save persists an order.
func (r *TenantRepo) Save(ctx context.Context, order *domain.Order) error {
	return r.inner.Save(ctx, order)
}

// Get retrieves an order only if it belongs to the current tenant.
func (r *TenantRepo) Get(ctx context.Context, id string) (*domain.Order, error) {
	order, err := r.inner.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil || order.TenantID != r.tenant(ctx) {
		return nil, nil
	}
	return order, nil
}

// ListByStatus lists orders for the current tenant.
func (r *TenantRepo) ListByStatus(ctx context.Context, tenantID, status string) ([]*domain.Order, error) {
	current := r.tenant(ctx)
	if tenantID != current {
		return nil, nil
	}
	return r.inner.ListByStatus(ctx, tenantID, status)
}
