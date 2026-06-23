package repo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/plantx/demo_app/backend/internal/domain"
)

// InMemoryRepo is a simple thread-safe in-memory repository.
type InMemoryRepo struct {
	mu        sync.RWMutex
	items     []*domain.Item
	settings  []*domain.Setting
	nextID    int
}

// NewInMemoryRepo creates a new InMemoryRepo.
func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{}
}

// CreateItem creates a new item scoped to the tenant.
func (r *InMemoryRepo) CreateItem(_ context.Context, tenantID, title string) (*domain.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	item := &domain.Item{
		ID:        uuid.NewString(),
		TenantID:  tenantID,
		Title:     title,
		CreatedAt: time.Now(),
	}
	r.items = append(r.items, item)
	return item, nil
}

// ListItems returns items for a tenant.
func (r *InMemoryRepo) ListItems(_ context.Context, tenantID string) ([]*domain.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Item
	for _, item := range r.items {
		if item.TenantID == tenantID {
			out = append(out, item)
		}
	}
	return out, nil
}

// CreateSetting creates a setting.
func (r *InMemoryRepo) CreateSetting(_ context.Context, setting *domain.Setting) (*domain.Setting, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.settings = append(r.settings, setting)
	return setting, nil
}

// GetSetting returns a setting by id.
func (r *InMemoryRepo) GetSetting(_ context.Context, id string) (*domain.Setting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.settings {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, fmt.Errorf("setting not found")
}

// ListSettings returns global settings plus settings for the given tenant.
func (r *InMemoryRepo) ListSettings(_ context.Context, tenantID string) ([]*domain.Setting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Setting
	for _, s := range r.settings {
		if s.Scope == domain.SettingScopeGlobal || s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

// UpdateSetting updates a setting value by id.
func (r *InMemoryRepo) UpdateSetting(_ context.Context, id, value string) (*domain.Setting, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.settings {
		if s.ID == id {
			s.Value = value
			return s, nil
		}
	}
	return nil, fmt.Errorf("setting not found")
}

// DeleteSetting removes a setting by id.
func (r *InMemoryRepo) DeleteSetting(_ context.Context, id string) (*domain.Setting, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, s := range r.settings {
		if s.ID == id {
			removed := s
			r.settings = append(r.settings[:i], r.settings[i+1:]...)
			return removed, nil
		}
	}
	return nil, fmt.Errorf("setting not found")
}
