package repo

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/plantx/platform/registry-service/internal/domain"
)

// InMemoryRepo is an in-memory implementation of domain.Repository for local
// development and testing. It is not durable and does not survive restarts.
type InMemoryRepo struct {
	mu           sync.RWMutex
	applications map[string]*domain.Application
	services     map[string]*domain.Service
	microApps    map[string]*domain.MicroApp
	menus        map[string]*domain.Menu
	policies     map[string]*domain.RoutePolicy
	menuOrder    int32
}

// NewInMemoryRepo creates a new InMemoryRepo.
func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		applications: make(map[string]*domain.Application),
		services:     make(map[string]*domain.Service),
		microApps:    make(map[string]*domain.MicroApp),
		menus:        make(map[string]*domain.Menu),
		policies:     make(map[string]*domain.RoutePolicy),
	}
}

// RegisterApplication registers a new application.
func (r *InMemoryRepo) RegisterApplication(_ context.Context, app *domain.Application) (*domain.Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if app == nil {
		return nil, nil
	}

	for _, existing := range r.applications {
		if existing.Key == app.Key {
			return nil, fmt.Errorf("application with key %q already exists", app.Key)
		}
	}

	stored := &domain.Application{
		ID:          uuid.NewString(),
		Key:         app.Key,
		Name:        app.Name,
		LabelKey:    app.LabelKey,
		Icon:        app.Icon,
		Description: app.Description,
		Status:      app.Status,
		SortOrder:   app.SortOrder,
	}
	if stored.Status == "" {
		stored.Status = domain.ApplicationStatusActive
	}
	r.applications[stored.ID] = stored
	return stored, nil
}

// ListApplications returns all registered applications.
func (r *InMemoryRepo) ListApplications(_ context.Context) ([]*domain.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.Application, 0, len(r.applications))
	for _, app := range r.applications {
		out = append(out, app)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SortOrder != out[j].SortOrder {
			return out[i].SortOrder < out[j].SortOrder
		}
		return out[i].Key < out[j].Key
	})
	return out, nil
}

// GetApplication returns an application by id.
func (r *InMemoryRepo) GetApplication(_ context.Context, id string) (*domain.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.applications[id], nil
}

// UpdateApplication updates an existing application.
func (r *InMemoryRepo) UpdateApplication(_ context.Context, app *domain.Application) (*domain.Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if app == nil {
		return nil, nil
	}
	stored, ok := r.applications[app.ID]
	if !ok {
		return nil, fmt.Errorf("application %q not found", app.ID)
	}
	if app.Key != "" && app.Key != stored.Key {
		for _, existing := range r.applications {
			if existing.ID != app.ID && existing.Key == app.Key {
				return nil, fmt.Errorf("application with key %q already exists", app.Key)
			}
		}
		stored.Key = app.Key
	}
	if app.Name != "" {
		stored.Name = app.Name
	}
	if app.LabelKey != "" {
		stored.LabelKey = app.LabelKey
	}
	stored.Icon = app.Icon
	stored.Description = app.Description
	if app.Status != "" {
		stored.Status = app.Status
	}
	stored.SortOrder = app.SortOrder
	return stored, nil
}

// DeleteApplication removes an application by id.
func (r *InMemoryRepo) DeleteApplication(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.applications, id)
	for _, svc := range r.services {
		if svc.ApplicationID == id {
			svc.ApplicationID = ""
			svc.ApplicationKey = ""
		}
	}
	for _, app := range r.microApps {
		if app.ApplicationID == id {
			app.ApplicationID = ""
			app.ApplicationKey = ""
		}
	}
	for _, m := range r.menus {
		if m.ApplicationID == id {
			m.ApplicationID = ""
			m.ApplicationKey = ""
		}
	}
	return nil
}

// GetApplicationMenus returns menus belonging to an application.
func (r *InMemoryRepo) GetApplicationMenus(_ context.Context, applicationID string) ([]*domain.Menu, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.Menu, 0)
	for _, m := range r.menus {
		if m.ApplicationID == applicationID {
			out = append(out, m)
		}
	}
	sortMenus(out)
	return out, nil
}

// GetApplicationMicroApps returns micro-apps belonging to an application.
func (r *InMemoryRepo) GetApplicationMicroApps(_ context.Context, applicationID string) ([]*domain.MicroApp, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.MicroApp, 0)
	for _, app := range r.microApps {
		if app.ApplicationID == applicationID {
			out = append(out, app)
		}
	}
	return out, nil
}

// RegisterService registers a service, replacing any existing entry with the same name.
func (r *InMemoryRepo) RegisterService(_ context.Context, name, grpcHost, restPrefix, applicationID string) (*domain.Service, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, svc := range r.services {
		if svc.Name == name {
			svc.GrpcHost = grpcHost
			svc.RestPrefix = restPrefix
			svc.Routes = []*domain.Route{{Path: restPrefix, Method: "*"}}
			svc.ApplicationID = applicationID
			svc.ApplicationKey = r.applicationKeyLocked(applicationID)
			return svc, nil
		}
	}

	svc := &domain.Service{
		ID:             uuid.NewString(),
		Name:           name,
		GrpcHost:       grpcHost,
		RestPrefix:     restPrefix,
		Routes:         []*domain.Route{{Path: restPrefix, Method: "*"}},
		ApplicationID:  applicationID,
		ApplicationKey: r.applicationKeyLocked(applicationID),
	}
	r.services[svc.ID] = svc
	return svc, nil
}

// DeregisterService removes a service by ID.
func (r *InMemoryRepo) DeregisterService(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	svc, ok := r.services[id]
	if !ok {
		return nil
	}
	delete(r.services, id)
	delete(r.policies, id)
	for _, app := range svc.MicroApps {
		delete(r.microApps, app.Name)
	}
	return nil
}

// GetService returns a service by ID.
func (r *InMemoryRepo) GetService(_ context.Context, id string) (*domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.services[id], nil
}

// ListServices returns all registered services.
func (r *InMemoryRepo) ListServices(_ context.Context) ([]*domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.Service, 0, len(r.services))
	for _, svc := range r.services {
		out = append(out, svc)
	}
	return out, nil
}

// RegisterMicroApp registers a micro-app manifest for a service by name.
func (r *InMemoryRepo) RegisterMicroApp(_ context.Context, serviceName string, microApp *domain.MicroApp) (*domain.MicroApp, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if microApp == nil {
		return nil, nil
	}

	svc, ok := r.findServiceByName(serviceName)
	if !ok {
		return nil, fmt.Errorf("service %q not found", serviceName)
	}

	app := &domain.MicroApp{
		Name:              microApp.Name,
		Route:             microApp.Route,
		BundleURL:         microApp.BundleURL,
		MenuLabelKey:      microApp.MenuLabelKey,
		RequirePermission: microApp.RequirePermission,
	}
	if app.Name == "" {
		app.Name = serviceName
	}
	if microApp.ApplicationID != "" {
		app.ApplicationID = microApp.ApplicationID
		app.ApplicationKey = r.applicationKeyLocked(microApp.ApplicationID)
	} else {
		app.ApplicationID = svc.ApplicationID
		app.ApplicationKey = svc.ApplicationKey
	}

	// Update existing micro-app in service if same name.
	found := false
	for _, existing := range svc.MicroApps {
		if existing.Name == app.Name {
			existing.Route = app.Route
			existing.BundleURL = app.BundleURL
			existing.MenuLabelKey = app.MenuLabelKey
			existing.RequirePermission = app.RequirePermission
			existing.ApplicationID = app.ApplicationID
			existing.ApplicationKey = app.ApplicationKey
			found = true
			break
		}
	}
	if !found {
		svc.MicroApps = append(svc.MicroApps, app)
	}
	r.microApps[app.Name] = app
	return app, nil
}

// ListMicroApps returns all registered micro-app manifests.
func (r *InMemoryRepo) ListMicroApps(_ context.Context) ([]*domain.MicroApp, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.MicroApp, 0, len(r.microApps))
	for _, app := range r.microApps {
		out = append(out, app)
	}
	return out, nil
}

// UpdateMicroApp updates a micro-app manifest by name.
func (r *InMemoryRepo) UpdateMicroApp(_ context.Context, name string, microApp *domain.MicroApp) (*domain.MicroApp, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if microApp == nil {
		return nil, nil
	}
	app, ok := r.microApps[name]
	if !ok {
		return nil, fmt.Errorf("micro-app %q not found", name)
	}
	app.Route = microApp.Route
	app.BundleURL = microApp.BundleURL
	app.MenuLabelKey = microApp.MenuLabelKey
	app.RequirePermission = microApp.RequirePermission
	return app, nil
}

// DeleteMicroApp removes a micro-app manifest by name.
func (r *InMemoryRepo) DeleteMicroApp(_ context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	app, ok := r.microApps[name]
	if !ok {
		return nil
	}
	delete(r.microApps, name)
	for _, svc := range r.services {
		updated := make([]*domain.MicroApp, 0, len(svc.MicroApps))
		for _, m := range svc.MicroApps {
			if m.Name != app.Name {
				updated = append(updated, m)
			}
		}
		svc.MicroApps = updated
	}
	return nil
}

// CreateMenu creates a new menu item.
func (r *InMemoryRepo) CreateMenu(_ context.Context, menu *domain.Menu) (*domain.Menu, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if menu == nil {
		return nil, nil
	}
	m := &domain.Menu{
		ID:                uuid.NewString(),
		LabelKey:          menu.LabelKey,
		Route:             menu.Route,
		Icon:              menu.Icon,
		ParentID:          menu.ParentID,
		SortOrder:         menu.SortOrder,
		MicroAppName:      menu.MicroAppName,
		RequirePermission: menu.RequirePermission,
		ApplicationID:     menu.ApplicationID,
		ApplicationKey:    r.applicationKeyLocked(menu.ApplicationID),
	}
	if m.SortOrder == 0 {
		r.menuOrder++
		m.SortOrder = r.menuOrder
	}
	r.menus[m.ID] = m
	return m, nil
}

// ListMenus returns all menu items ordered by parent and sort order.
func (r *InMemoryRepo) ListMenus(_ context.Context) ([]*domain.Menu, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.Menu, 0, len(r.menus))
	for _, m := range r.menus {
		out = append(out, m)
	}
	sortMenus(out)
	return out, nil
}

// UpdateMenu updates a menu item by id.
func (r *InMemoryRepo) UpdateMenu(_ context.Context, menu *domain.Menu) (*domain.Menu, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if menu == nil {
		return nil, nil
	}
	m, ok := r.menus[menu.ID]
	if !ok {
		return nil, fmt.Errorf("menu %q not found", menu.ID)
	}
	m.LabelKey = menu.LabelKey
	m.Route = menu.Route
	m.Icon = menu.Icon
	m.ParentID = menu.ParentID
	m.SortOrder = menu.SortOrder
	m.MicroAppName = menu.MicroAppName
	m.RequirePermission = menu.RequirePermission
	if menu.ApplicationID != "" {
		m.ApplicationID = menu.ApplicationID
		m.ApplicationKey = r.applicationKeyLocked(menu.ApplicationID)
	}
	return m, nil
}

// DeleteMenu removes a menu item by id.
func (r *InMemoryRepo) DeleteMenu(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.menus[id]; !ok {
		return nil
	}
	delete(r.menus, id)
	for _, m := range r.menus {
		if m.ParentID == id {
			m.ParentID = ""
		}
	}
	return nil
}

// ReorderMenus updates sort order for multiple menu items.
func (r *InMemoryRepo) ReorderMenus(_ context.Context, order map[string]int32) ([]*domain.Menu, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, sortOrder := range order {
		if m, ok := r.menus[id]; ok {
			m.SortOrder = sortOrder
		}
	}
	out := make([]*domain.Menu, 0, len(r.menus))
	for _, m := range r.menus {
		out = append(out, m)
	}
	sortMenus(out)
	return out, nil
}

func sortMenus(menus []*domain.Menu) {
	sort.SliceStable(menus, func(i, j int) bool {
		pi, pj := menus[i].ParentID, menus[j].ParentID
		if pi == "" && pj != "" {
			return true
		}
		if pi != "" && pj == "" {
			return false
		}
		if pi != pj {
			return pi < pj
		}
		if menus[i].SortOrder != menus[j].SortOrder {
			return menus[i].SortOrder < menus[j].SortOrder
		}
		return menus[i].LabelKey < menus[j].LabelKey
	})
}

// GetRoutePolicy returns the route policy for a service.
func (r *InMemoryRepo) GetRoutePolicy(_ context.Context, serviceID string) (*domain.RoutePolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if policy, ok := r.policies[serviceID]; ok {
		return policy, nil
	}
	return nil, nil
}

// SetRoutePolicy stores the route policy for a service.
func (r *InMemoryRepo) SetRoutePolicy(_ context.Context, serviceID string, policy *domain.RoutePolicy) (*domain.RoutePolicy, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.services[serviceID]; !ok {
		return nil, fmt.Errorf("service %q not found", serviceID)
	}
	if policy == nil {
		policy = &domain.RoutePolicy{AuthRequired: true}
	}
	stored := &domain.RoutePolicy{
		RateLimitRPS: policy.RateLimitRPS,
		AuthRequired: policy.AuthRequired,
		CanaryWeight: policy.CanaryWeight,
		CanaryHost:   policy.CanaryHost,
	}
	r.policies[serviceID] = stored
	return stored, nil
}

// DeleteRoutePolicy removes the route policy for a service.
func (r *InMemoryRepo) DeleteRoutePolicy(_ context.Context, serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.policies, serviceID)
	return nil
}

// ListRoutePolicies returns all route policies keyed by service id.
func (r *InMemoryRepo) ListRoutePolicies(_ context.Context) (map[string]*domain.RoutePolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string]*domain.RoutePolicy, len(r.policies))
	for id, policy := range r.policies {
		out[id] = policy
	}
	return out, nil
}

func (r *InMemoryRepo) findServiceByName(name string) (*domain.Service, bool) {
	for _, svc := range r.services {
		if svc.Name == name {
			return svc, true
		}
	}
	return nil, false
}

func (r *InMemoryRepo) applicationKeyLocked(id string) string {
	if app, ok := r.applications[id]; ok {
		return app.Key
	}
	return ""
}
