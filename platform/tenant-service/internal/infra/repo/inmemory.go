package repo

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/plantx/platform/tenant-service/internal/domain"
)

// InMemoryRepo is an in-memory tenant repository.
type InMemoryRepo struct {
	mu      sync.RWMutex
	tenants map[string]*domain.Tenant
}

// NewInMemoryRepo creates a new InMemoryRepo.
func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{tenants: make(map[string]*domain.Tenant)}
}

// Create stores a new tenant.
func (r *InMemoryRepo) Create(_ context.Context, name string) (*domain.Tenant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := &domain.Tenant{
		ID:        uuid.NewString(),
		Name:      name,
		Status:    "active",
		CreatedAt: time.Now().Unix(),
	}
	r.tenants[t.ID] = t
	return t, nil
}

// List returns all tenants.
func (r *InMemoryRepo) List(_ context.Context) ([]*domain.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.Tenant, 0, len(r.tenants))
	for _, t := range r.tenants {
		out = append(out, t)
	}
	return out, nil
}
