package app

import (
	"context"

	"github.com/plantx/platform/tenant-service/internal/domain"
)

// TenantService implements tenant use cases.
type TenantService struct {
	repo domain.Repository
}

// NewTenantService creates a new TenantService.
func NewTenantService(repo domain.Repository) *TenantService {
	return &TenantService{repo: repo}
}

// CreateTenant creates a new tenant.
func (s *TenantService) CreateTenant(ctx context.Context, name string) (*domain.Tenant, error) {
	return s.repo.Create(ctx, name)
}

// ListTenants lists all tenants.
func (s *TenantService) ListTenants(ctx context.Context) ([]*domain.Tenant, error) {
	return s.repo.List(ctx)
}
