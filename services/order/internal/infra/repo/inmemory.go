package repo

import (
	"context"
	"sync"

	"github.com/plantx/services/order/internal/domain"
)

// InMemoryRepo is a temporary in-memory repository for M1 demos.
type InMemoryRepo struct {
	mu     sync.RWMutex
	orders map[string]*domain.Order
}

// NewInMemoryRepo creates a new InMemoryRepo.
func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{orders: make(map[string]*domain.Order)}
}

// Save stores an order.
func (r *InMemoryRepo) Save(ctx context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
	return nil
}

// Get retrieves an order by id.
func (r *InMemoryRepo) Get(ctx context.Context, id string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if o, ok := r.orders[id]; ok {
		return o, nil
	}
	return nil, nil
}

// ListByStatus lists orders for a tenant filtered by status.
func (r *InMemoryRepo) ListByStatus(ctx context.Context, tenantID, status string) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Order
	for _, o := range r.orders {
		if o.TenantID == tenantID && o.Status == status {
			result = append(result, o)
		}
	}
	return result, nil
}
