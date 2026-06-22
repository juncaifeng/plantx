package app

import (
	"context"

	"github.com/plantx/platform/gateway-service/internal/domain"
)

// Registry implements gateway registry use cases.
type Registry struct {
	repo domain.Repository
}

// NewRegistry creates a new Registry.
func NewRegistry(repo domain.Repository) *Registry {
	return &Registry{repo: repo}
}

// RegisterService registers a backend service.
func (r *Registry) RegisterService(ctx context.Context, name, grpcHost, restPrefix string) (*domain.Service, error) {
	return r.repo.Register(ctx, name, grpcHost, restPrefix, nil)
}

// ListServices lists registered services.
func (r *Registry) ListServices(ctx context.Context) ([]*domain.Service, error) {
	return r.repo.List(ctx)
}

// GetRoutes returns routes for a service.
func (r *Registry) GetRoutes(ctx context.Context, id string) ([]*domain.Route, bool, error) {
	return r.repo.Routes(ctx, id)
}

// RegisterMicroApp registers a micro-app manifest.
func (r *Registry) RegisterMicroApp(ctx context.Context, serviceName string, app *domain.MicroApp) (*domain.MicroApp, error) {
	return r.repo.RegisterMicroApp(ctx, serviceName, app)
}

// ListMicroApps lists registered micro-app manifests.
func (r *Registry) ListMicroApps(ctx context.Context) ([]*domain.MicroApp, error) {
	return r.repo.ListMicroApps(ctx)
}
