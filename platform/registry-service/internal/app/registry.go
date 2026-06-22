package app

import (
	"context"

	"github.com/plantx/platform/registry-service/internal/domain"
)

// Registry implements registry use cases.
type Registry struct {
	repo domain.Repository
}

// NewRegistry creates a new Registry application service.
func NewRegistry(repo domain.Repository) *Registry {
	return &Registry{repo: repo}
}

// RegisterApplication registers a new application.
func (r *Registry) RegisterApplication(ctx context.Context, app *domain.Application) (*domain.Application, error) {
	return r.repo.RegisterApplication(ctx, app)
}

// ListApplications returns all applications.
func (r *Registry) ListApplications(ctx context.Context) ([]*domain.Application, error) {
	return r.repo.ListApplications(ctx)
}

// GetApplication returns an application by id.
func (r *Registry) GetApplication(ctx context.Context, id string) (*domain.Application, error) {
	return r.repo.GetApplication(ctx, id)
}

// UpdateApplication updates an application.
func (r *Registry) UpdateApplication(ctx context.Context, app *domain.Application) (*domain.Application, error) {
	return r.repo.UpdateApplication(ctx, app)
}

// DeleteApplication removes an application.
func (r *Registry) DeleteApplication(ctx context.Context, id string) error {
	return r.repo.DeleteApplication(ctx, id)
}

// GetApplicationMenus returns menus for an application.
func (r *Registry) GetApplicationMenus(ctx context.Context, applicationID string) ([]*domain.Menu, error) {
	return r.repo.GetApplicationMenus(ctx, applicationID)
}

// GetApplicationMicroApps returns micro-apps for an application.
func (r *Registry) GetApplicationMicroApps(ctx context.Context, applicationID string) ([]*domain.MicroApp, error) {
	return r.repo.GetApplicationMicroApps(ctx, applicationID)
}

// RegisterService registers a backend service.
func (r *Registry) RegisterService(ctx context.Context, name, grpcHost, restPrefix, applicationID string) (*domain.Service, error) {
	return r.repo.RegisterService(ctx, name, grpcHost, restPrefix, applicationID)
}

// DeregisterService removes a service by id.
func (r *Registry) DeregisterService(ctx context.Context, id string) error {
	return r.repo.DeregisterService(ctx, id)
}

// GetService returns a service by id.
func (r *Registry) GetService(ctx context.Context, id string) (*domain.Service, error) {
	return r.repo.GetService(ctx, id)
}

// ListServices returns all registered services.
func (r *Registry) ListServices(ctx context.Context) ([]*domain.Service, error) {
	return r.repo.ListServices(ctx)
}

// RegisterMicroApp registers a micro-app manifest.
func (r *Registry) RegisterMicroApp(ctx context.Context, serviceName string, microApp *domain.MicroApp) (*domain.MicroApp, error) {
	return r.repo.RegisterMicroApp(ctx, serviceName, microApp)
}

// ListMicroApps returns all registered micro-app manifests.
func (r *Registry) ListMicroApps(ctx context.Context) ([]*domain.MicroApp, error) {
	return r.repo.ListMicroApps(ctx)
}

// UpdateMicroApp updates a micro-app manifest.
func (r *Registry) UpdateMicroApp(ctx context.Context, name string, microApp *domain.MicroApp) (*domain.MicroApp, error) {
	return r.repo.UpdateMicroApp(ctx, name, microApp)
}

// DeleteMicroApp removes a micro-app manifest.
func (r *Registry) DeleteMicroApp(ctx context.Context, name string) error {
	return r.repo.DeleteMicroApp(ctx, name)
}

// CreateMenu creates a new menu item.
func (r *Registry) CreateMenu(ctx context.Context, menu *domain.Menu) (*domain.Menu, error) {
	return r.repo.CreateMenu(ctx, menu)
}

// ListMenus returns all menu items.
func (r *Registry) ListMenus(ctx context.Context) ([]*domain.Menu, error) {
	return r.repo.ListMenus(ctx)
}

// UpdateMenu updates a menu item.
func (r *Registry) UpdateMenu(ctx context.Context, menu *domain.Menu) (*domain.Menu, error) {
	return r.repo.UpdateMenu(ctx, menu)
}

// DeleteMenu removes a menu item.
func (r *Registry) DeleteMenu(ctx context.Context, id string) error {
	return r.repo.DeleteMenu(ctx, id)
}

// ReorderMenus updates sort order for menu items.
func (r *Registry) ReorderMenus(ctx context.Context, order map[string]int32) ([]*domain.Menu, error) {
	return r.repo.ReorderMenus(ctx, order)
}

// GetRoutePolicy returns the gateway route policy for a service.
func (r *Registry) GetRoutePolicy(ctx context.Context, serviceID string) (*domain.RoutePolicy, error) {
	return r.repo.GetRoutePolicy(ctx, serviceID)
}

// SetRoutePolicy updates the gateway route policy for a service.
func (r *Registry) SetRoutePolicy(ctx context.Context, serviceID string, policy *domain.RoutePolicy) (*domain.RoutePolicy, error) {
	return r.repo.SetRoutePolicy(ctx, serviceID, policy)
}

// SyncRoutes returns the full route manifest for the gateway.
func (r *Registry) SyncRoutes(ctx context.Context) ([]*domain.ServiceRoute, error) {
	services, err := r.repo.ListServices(ctx)
	if err != nil {
		return nil, err
	}

	routes := make([]*domain.ServiceRoute, 0, len(services))
	for _, svc := range services {
		routes = append(routes, &domain.ServiceRoute{
			ServiceID:    svc.ID,
			Name:         svc.Name,
			RestPrefix:   svc.RestPrefix,
			UpstreamHost: svc.Name + ":8081",
			Routes:       svc.Routes,
			Policy:       svc.Policy,
		})
	}
	return routes, nil
}
