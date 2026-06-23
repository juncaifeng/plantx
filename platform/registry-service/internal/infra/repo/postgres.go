package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/plantx/kit/kit-go/db"
	"github.com/plantx/platform/registry-service/internal/domain"
	"github.com/plantx/platform/registry-service/internal/infra/sqlc"
)

// PostgresRepo implements domain.Repository using PostgreSQL via sqlc.
type PostgresRepo struct {
	queries *sqlc.Queries
}

// NewPostgresRepo creates a new PostgresRepo.
func NewPostgresRepo(d db.DB) *PostgresRepo {
	return &PostgresRepo{queries: sqlc.New(d)}
}

func (r *PostgresRepo) applicationKeyMap(ctx context.Context) (map[uuid.UUID]string, error) {
	apps, err := r.queries.ListApplications(ctx)
	if err != nil {
		return nil, fmt.Errorf("load applications: %w", err)
	}
	m := make(map[uuid.UUID]string, len(apps))
	for _, app := range apps {
		m[app.ID] = app.Key
	}
	return m, nil
}

// RegisterApplication persists a new application.
func (r *PostgresRepo) RegisterApplication(ctx context.Context, app *domain.Application) (*domain.Application, error) {
	if app == nil {
		return nil, nil
	}
	row, err := r.queries.CreateApplication(ctx, sqlc.CreateApplicationParams{
		Key:         app.Key,
		Name:        app.Name,
		LabelKey:    app.LabelKey,
		Icon:        app.Icon,
		Description: app.Description,
		Status:      string(app.Status),
		SortOrder:   app.SortOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}
	return toDomainApplication(row), nil
}

// ListApplications returns all applications ordered by sort_order and key.
func (r *PostgresRepo) ListApplications(ctx context.Context) ([]*domain.Application, error) {
	rows, err := r.queries.ListApplications(ctx)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	apps := make([]*domain.Application, 0, len(rows))
	for _, row := range rows {
		apps = append(apps, toDomainApplication(row))
	}
	return apps, nil
}

// GetApplication returns an application by id.
func (r *PostgresRepo) GetApplication(ctx context.Context, id string) (*domain.Application, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid application id: %w", err)
	}
	row, err := r.queries.GetApplication(ctx, uid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get application: %w", err)
	}
	return toDomainApplication(row), nil
}

// UpdateApplication updates an existing application.
func (r *PostgresRepo) UpdateApplication(ctx context.Context, app *domain.Application) (*domain.Application, error) {
	if app == nil {
		return nil, nil
	}
	uid, err := uuid.Parse(app.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid application id: %w", err)
	}
	row, err := r.queries.UpdateApplication(ctx, sqlc.UpdateApplicationParams{
		ID:          uid,
		Key:         app.Key,
		Name:        app.Name,
		LabelKey:    app.LabelKey,
		Icon:        app.Icon,
		Description: app.Description,
		Status:      string(app.Status),
		SortOrder:   app.SortOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("update application: %w", err)
	}
	return toDomainApplication(row), nil
}

// DeleteApplication removes an application by id.
func (r *PostgresRepo) DeleteApplication(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid application id: %w", err)
	}
	if err := r.queries.DeleteApplication(ctx, uid); err != nil {
		return fmt.Errorf("delete application: %w", err)
	}
	return nil
}

// GetApplicationMenus returns menus belonging to an application.
func (r *PostgresRepo) GetApplicationMenus(ctx context.Context, applicationID string) ([]*domain.Menu, error) {
	uid, err := uuid.Parse(applicationID)
	if err != nil {
		return nil, fmt.Errorf("invalid application id: %w", err)
	}
	rows, err := r.queries.ListMenusByApplicationID(ctx, uuid.NullUUID{UUID: uid, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list application menus: %w", err)
	}
	menus := make([]*domain.Menu, 0, len(rows))
	for _, row := range rows {
		menus = append(menus, toDomainMenu(row))
	}
	return menus, nil
}

// GetApplicationMicroApps returns micro-apps belonging to an application.
func (r *PostgresRepo) GetApplicationMicroApps(ctx context.Context, applicationID string) ([]*domain.MicroApp, error) {
	uid, err := uuid.Parse(applicationID)
	if err != nil {
		return nil, fmt.Errorf("invalid application id: %w", err)
	}
	rows, err := r.queries.ListMicroAppsByApplicationID(ctx, uuid.NullUUID{UUID: uid, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list application micro-apps: %w", err)
	}
	apps := make([]*domain.MicroApp, 0, len(rows))
	for _, row := range rows {
		apps = append(apps, toDomainMicroApp(row))
	}
	return apps, nil
}

// RegisterService persists a service, upserting by name.
func (r *PostgresRepo) RegisterService(ctx context.Context, name, grpcHost, restPrefix, applicationID string) (*domain.Service, error) {
	appID := nullUUID(applicationID)
	if applicationID == "" {
		// Default to the platform application for backwards compatibility.
		if app, err := r.queries.GetApplicationByKey(ctx, "platform"); err == nil {
			appID = uuid.NullUUID{UUID: app.ID, Valid: true}
		}
	}
	svc, err := r.queries.UpsertService(ctx, sqlc.UpsertServiceParams{
		Name:          name,
		GrpcHost:      grpcHost,
		RestPrefix:    restPrefix,
		ApplicationID: appID,
		Status:        string(domain.ResourceStatusOnline),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert service: %w", err)
	}
	return r.serviceToDomain(ctx, svc)
}

// DeregisterService removes a service by ID.
func (r *PostgresRepo) DeregisterService(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid service id: %w", err)
	}
	if err := r.queries.DeleteService(ctx, uid); err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	return nil
}

// UpdateServiceStatus updates the status of a service by name.
func (r *PostgresRepo) UpdateServiceStatus(ctx context.Context, name string, status domain.ResourceStatus) (*domain.Service, error) {
	svc, err := r.queries.UpdateServiceStatusByName(ctx, sqlc.UpdateServiceStatusByNameParams{
		Name:   name,
		Status: string(status),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update service status: %w", err)
	}
	return r.serviceToDomain(ctx, svc)
}

// GetService returns a service by ID.
func (r *PostgresRepo) GetService(ctx context.Context, id string) (*domain.Service, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid service id: %w", err)
	}
	svc, err := r.queries.GetService(ctx, uid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get service: %w", err)
	}
	return r.serviceToDomain(ctx, svc)
}

// ListServices returns all registered services with their micro-apps.
func (r *PostgresRepo) ListServices(ctx context.Context) ([]*domain.Service, error) {
	rows, err := r.queries.ListServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	services := make([]*domain.Service, 0, len(rows))
	for _, svc := range rows {
		d, err := r.serviceToDomain(ctx, svc)
		if err != nil {
			return nil, err
		}
		services = append(services, d)
	}
	return services, nil
}

// RegisterMicroApp attaches or updates a micro-app manifest for a service by name.
func (r *PostgresRepo) RegisterMicroApp(ctx context.Context, serviceName string, microApp *domain.MicroApp) (*domain.MicroApp, error) {
	if microApp == nil {
		return nil, nil
	}
	svc, err := r.queries.GetServiceByName(ctx, serviceName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("service %q not found", serviceName)
	}
	if err != nil {
		return nil, fmt.Errorf("lookup service %q: %w", serviceName, err)
	}

	name := microApp.Name
	if name == "" {
		name = serviceName
	}

	appID := nullUUID(microApp.ApplicationID)
	if !appID.Valid && svc.ApplicationID.Valid {
		appID = svc.ApplicationID
	}

	row, err := r.queries.UpsertMicroApp(ctx, sqlc.UpsertMicroAppParams{
		ServiceID:         svc.ID,
		Name:              name,
		Route:             microApp.Route,
		BundleUrl:         microApp.BundleURL,
		MenuLabelKey:      microApp.MenuLabelKey,
		RequirePermission: microApp.RequirePermission,
		ApplicationID:     appID,
		Upstream:          nullString(microApp.Upstream),
		Status:            defaultResourceStatus(microApp.Status),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert micro-app: %w", err)
	}
	return toDomainMicroApp(row), nil
}

// ListMicroApps returns all registered micro-app manifests.
func (r *PostgresRepo) ListMicroApps(ctx context.Context) ([]*domain.MicroApp, error) {
	rows, err := r.queries.ListMicroApps(ctx)
	if err != nil {
		return nil, fmt.Errorf("list micro-apps: %w", err)
	}
	keys, err := r.applicationKeyMap(ctx)
	if err != nil {
		return nil, err
	}
	apps := make([]*domain.MicroApp, 0, len(rows))
	for _, row := range rows {
		app := toDomainMicroApp(row)
		if row.ApplicationID.Valid {
			app.ApplicationKey = keys[row.ApplicationID.UUID]
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// UpdateMicroApp updates a micro-app manifest by name.
func (r *PostgresRepo) UpdateMicroApp(ctx context.Context, name string, microApp *domain.MicroApp) (*domain.MicroApp, error) {
	if microApp == nil {
		return nil, nil
	}
	row, err := r.queries.UpdateMicroApp(ctx, sqlc.UpdateMicroAppParams{
		Name:              name,
		Route:             microApp.Route,
		BundleUrl:         microApp.BundleURL,
		MenuLabelKey:      microApp.MenuLabelKey,
		RequirePermission: microApp.RequirePermission,
		Upstream:          nullString(microApp.Upstream),
		Status:            defaultResourceStatus(microApp.Status),
	})
	if err != nil {
		return nil, fmt.Errorf("update micro-app: %w", err)
	}
	return toDomainMicroApp(row), nil
}

// DeleteMicroApp removes a micro-app manifest by name.
func (r *PostgresRepo) DeleteMicroApp(ctx context.Context, name string) error {
	if err := r.queries.DeleteMicroApp(ctx, name); err != nil {
		return fmt.Errorf("delete micro-app: %w", err)
	}
	return nil
}

// CreateMenu creates a new menu item.
func (r *PostgresRepo) CreateMenu(ctx context.Context, menu *domain.Menu) (*domain.Menu, error) {
	if menu == nil {
		return nil, nil
	}
	row, err := r.queries.CreateMenu(ctx, sqlc.CreateMenuParams{
		LabelKey:          menu.LabelKey,
		Route:             menu.Route,
		Icon:              menu.Icon,
		ParentID:          nullUUID(menu.ParentID),
		SortOrder:         menu.SortOrder,
		MicroAppName:      menu.MicroAppName,
		RequirePermission: menu.RequirePermission,
		ApplicationID:     nullUUID(menu.ApplicationID),
		Status:            defaultResourceStatus(menu.Status),
	})
	if err != nil {
		return nil, fmt.Errorf("create menu: %w", err)
	}
	return toDomainMenu(row), nil
}

// ListMenus returns all menu items.
func (r *PostgresRepo) ListMenus(ctx context.Context) ([]*domain.Menu, error) {
	rows, err := r.queries.ListMenus(ctx)
	if err != nil {
		return nil, fmt.Errorf("list menus: %w", err)
	}
	keys, err := r.applicationKeyMap(ctx)
	if err != nil {
		return nil, err
	}
	menus := make([]*domain.Menu, 0, len(rows))
	for _, row := range rows {
		menu := toDomainMenu(row)
		if row.ApplicationID.Valid {
			menu.ApplicationKey = keys[row.ApplicationID.UUID]
		}
		menus = append(menus, menu)
	}
	return menus, nil
}

// UpdateMenu updates a menu item by id.
func (r *PostgresRepo) UpdateMenu(ctx context.Context, menu *domain.Menu) (*domain.Menu, error) {
	if menu == nil {
		return nil, nil
	}
	uid, err := uuid.Parse(menu.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid menu id: %w", err)
	}
	row, err := r.queries.UpdateMenu(ctx, sqlc.UpdateMenuParams{
		ID:                uid,
		LabelKey:          menu.LabelKey,
		Route:             menu.Route,
		Icon:              menu.Icon,
		ParentID:          nullUUID(menu.ParentID),
		SortOrder:         menu.SortOrder,
		MicroAppName:      menu.MicroAppName,
		RequirePermission: menu.RequirePermission,
		ApplicationID:     nullUUID(menu.ApplicationID),
		Status:            defaultResourceStatus(menu.Status),
	})
	if err != nil {
		return nil, fmt.Errorf("update menu: %w", err)
	}
	return toDomainMenu(row), nil
}

// DeleteMenu removes a menu item by id.
func (r *PostgresRepo) DeleteMenu(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid menu id: %w", err)
	}
	if err := r.queries.DeleteMenu(ctx, uid); err != nil {
		return fmt.Errorf("delete menu: %w", err)
	}
	return nil
}

// ReorderMenus updates sort order for multiple menu items.
func (r *PostgresRepo) ReorderMenus(ctx context.Context, order map[string]int32) ([]*domain.Menu, error) {
	for id, sortOrder := range order {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid menu id %q: %w", id, err)
		}
		if _, err := r.queries.UpdateMenuSortOrder(ctx, sqlc.UpdateMenuSortOrderParams{
			ID:        uid,
			SortOrder: sortOrder,
		}); err != nil {
			return nil, fmt.Errorf("reorder menu %q: %w", id, err)
		}
	}
	return r.ListMenus(ctx)
}

func (r *PostgresRepo) serviceToDomain(ctx context.Context, svc sqlc.RegistryService) (*domain.Service, error) {
	keys, err := r.applicationKeyMap(ctx)
	if err != nil {
		return nil, err
	}
	microApps, err := r.microAppsForService(ctx, svc.ID, keys)
	if err != nil {
		return nil, err
	}
	policy, err := r.routePolicyForService(ctx, svc.ID)
	if err != nil {
		return nil, err
	}
	d := &domain.Service{
		ID:         svc.ID.String(),
		Name:       svc.Name,
		GrpcHost:   svc.GrpcHost,
		RestPrefix: svc.RestPrefix,
		Routes: []*domain.Route{
			{Path: svc.RestPrefix, Method: "*"},
		},
		MicroApps: microApps,
		Policy:    policy,
		Status:    domain.ResourceStatus(svc.Status),
	}
	if svc.ApplicationID.Valid {
		d.ApplicationID = svc.ApplicationID.UUID.String()
		d.ApplicationKey = keys[svc.ApplicationID.UUID]
	}
	return d, nil
}

func (r *PostgresRepo) microAppsForService(ctx context.Context, serviceID uuid.UUID, keys map[uuid.UUID]string) ([]*domain.MicroApp, error) {
	rows, err := r.queries.ListMicroAppsByServiceID(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("list micro-apps for service: %w", err)
	}
	apps := make([]*domain.MicroApp, 0, len(rows))
	for _, row := range rows {
		app := toDomainMicroApp(row)
		if row.ApplicationID.Valid {
			app.ApplicationKey = keys[row.ApplicationID.UUID]
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func (r *PostgresRepo) routePolicyForService(ctx context.Context, serviceID uuid.UUID) (*domain.RoutePolicy, error) {
	row, err := r.queries.GetRoutePolicy(ctx, serviceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get route policy for service: %w", err)
	}
	return toDomainRoutePolicy(row), nil
}

// GetRoutePolicy returns the route policy for a service.
func (r *PostgresRepo) GetRoutePolicy(ctx context.Context, serviceID string) (*domain.RoutePolicy, error) {
	uid, err := uuid.Parse(serviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid service id: %w", err)
	}
	return r.routePolicyForService(ctx, uid)
}

// SetRoutePolicy creates or updates the route policy for a service.
func (r *PostgresRepo) SetRoutePolicy(ctx context.Context, serviceID string, policy *domain.RoutePolicy) (*domain.RoutePolicy, error) {
	uid, err := uuid.Parse(serviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid service id: %w", err)
	}
	if policy == nil {
		policy = &domain.RoutePolicy{AuthRequired: true}
	}
	row, err := r.queries.UpsertRoutePolicy(ctx, sqlc.UpsertRoutePolicyParams{
		ServiceID:    uid,
		RateLimitRps: policy.RateLimitRPS,
		AuthRequired: policy.AuthRequired,
		CanaryWeight: policy.CanaryWeight,
		CanaryHost:   nullString(policy.CanaryHost),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert route policy: %w", err)
	}
	return toDomainRoutePolicy(row), nil
}

// DeleteRoutePolicy removes the route policy for a service.
func (r *PostgresRepo) DeleteRoutePolicy(ctx context.Context, serviceID string) error {
	uid, err := uuid.Parse(serviceID)
	if err != nil {
		return fmt.Errorf("invalid service id: %w", err)
	}
	if err := r.queries.DeleteRoutePolicy(ctx, uid); err != nil {
		return fmt.Errorf("delete route policy: %w", err)
	}
	return nil
}

// ListRoutePolicies returns all route policies keyed by service id.
func (r *PostgresRepo) ListRoutePolicies(ctx context.Context) (map[string]*domain.RoutePolicy, error) {
	rows, err := r.queries.ListRoutePolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list route policies: %w", err)
	}
	policies := make(map[string]*domain.RoutePolicy, len(rows))
	for _, row := range rows {
		policies[row.ServiceID.String()] = toDomainRoutePolicy(row)
	}
	return policies, nil
}

func toDomainApplication(row sqlc.Application) *domain.Application {
	return &domain.Application{
		ID:          row.ID.String(),
		Key:         row.Key,
		Name:        row.Name,
		LabelKey:    row.LabelKey,
		Icon:        row.Icon,
		Description: row.Description,
		Status:      domain.ApplicationStatus(row.Status),
		SortOrder:   row.SortOrder,
	}
}

func toDomainMicroApp(row sqlc.MicroApp) *domain.MicroApp {
	m := &domain.MicroApp{
		Name:              row.Name,
		Route:             row.Route,
		BundleURL:         row.BundleUrl,
		MenuLabelKey:      row.MenuLabelKey,
		RequirePermission: row.RequirePermission,
		Status:            domain.ResourceStatus(row.Status),
	}
	if row.ApplicationID.Valid {
		m.ApplicationID = row.ApplicationID.UUID.String()
	}
	if row.Upstream.Valid {
		m.Upstream = row.Upstream.String
	}
	return m
}

func toDomainMenu(row sqlc.Menu) *domain.Menu {
	m := &domain.Menu{
		ID:                row.ID.String(),
		LabelKey:          row.LabelKey,
		Route:             row.Route,
		Icon:              row.Icon,
		SortOrder:         row.SortOrder,
		MicroAppName:      row.MicroAppName,
		RequirePermission: row.RequirePermission,
		Status:            domain.ResourceStatus(row.Status),
	}
	if row.ParentID.Valid {
		m.ParentID = row.ParentID.UUID.String()
	}
	if row.ApplicationID.Valid {
		m.ApplicationID = row.ApplicationID.UUID.String()
	}
	return m
}

func toDomainRoutePolicy(row sqlc.RoutePolicy) *domain.RoutePolicy {
	p := &domain.RoutePolicy{
		RateLimitRPS: row.RateLimitRps,
		AuthRequired: row.AuthRequired,
		CanaryWeight: row.CanaryWeight,
	}
	if row.CanaryHost.Valid {
		p.CanaryHost = row.CanaryHost.String
	}
	return p
}

func defaultResourceStatus(s domain.ResourceStatus) string {
	if s == "" {
		return string(domain.ResourceStatusOnline)
	}
	return string(s)
}

func nullUUID(s string) uuid.NullUUID {
	if s == "" {
		return uuid.NullUUID{}
	}
	uid, err := uuid.Parse(s)
	if err != nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: uid, Valid: true}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
