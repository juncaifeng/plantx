package grpc

import (
	"context"

	"github.com/plantx/platform/registry-service/api"
	"github.com/plantx/platform/registry-service/internal/app"
	"github.com/plantx/platform/registry-service/internal/domain"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Handler implements the RegistryService gRPC server.
type Handler struct {
	api.UnimplementedRegistryServiceServer
	registry *app.Registry
}

// NewHandler creates a new Handler.
func NewHandler(registry *app.Registry) *Handler {
	return &Handler{registry: registry}
}

// RegisterApplication registers a new application.
func (h *Handler) RegisterApplication(ctx context.Context, req *api.RegisterApplicationRequest) (*api.Application, error) {
	app, err := h.registry.RegisterApplication(ctx, toDomainApplicationFromRegister(req))
	if err != nil {
		return nil, err
	}
	return toProtoApplication(app), nil
}

// ListApplications lists all applications.
func (h *Handler) ListApplications(ctx context.Context, _ *api.ListApplicationsRequest) (*api.ApplicationList, error) {
	apps, err := h.registry.ListApplications(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.Application, 0, len(apps))
	for _, a := range apps {
		out = append(out, toProtoApplication(a))
	}
	return &api.ApplicationList{Applications: out}, nil
}

// GetApplication returns an application by id.
func (h *Handler) GetApplication(ctx context.Context, req *api.GetApplicationRequest) (*api.Application, error) {
	app, err := h.registry.GetApplication(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return toProtoApplication(app), nil
}

// UpdateApplication updates an application.
func (h *Handler) UpdateApplication(ctx context.Context, req *api.UpdateApplicationRequest) (*api.Application, error) {
	app, err := h.registry.UpdateApplication(ctx, toDomainApplicationFromUpdate(req))
	if err != nil {
		return nil, err
	}
	return toProtoApplication(app), nil
}

// DeleteApplication removes an application.
func (h *Handler) DeleteApplication(ctx context.Context, req *api.DeleteApplicationRequest) (*emptypb.Empty, error) {
	if err := h.registry.DeleteApplication(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// GetApplicationMenus returns menus for an application.
func (h *Handler) GetApplicationMenus(ctx context.Context, req *api.GetApplicationMenusRequest) (*api.MenuList, error) {
	menus, err := h.registry.GetApplicationMenus(ctx, req.GetApplicationId())
	if err != nil {
		return nil, err
	}
	out := make([]*api.Menu, 0, len(menus))
	for _, m := range menus {
		out = append(out, toProtoMenu(m))
	}
	return &api.MenuList{Menus: out}, nil
}

// GetApplicationMicroApps returns micro-apps for an application.
func (h *Handler) GetApplicationMicroApps(ctx context.Context, req *api.GetApplicationMicroAppsRequest) (*api.MicroAppList, error) {
	apps, err := h.registry.GetApplicationMicroApps(ctx, req.GetApplicationId())
	if err != nil {
		return nil, err
	}
	out := make([]*api.MicroApp, 0, len(apps))
	for _, a := range apps {
		out = append(out, toProtoMicroApp(a))
	}
	return &api.MicroAppList{MicroApps: out}, nil
}

// RegisterService registers a backend service.
func (h *Handler) RegisterService(ctx context.Context, req *api.RegisterServiceRequest) (*api.Service, error) {
	svc, err := h.registry.RegisterService(ctx, req.GetName(), req.GetGrpcHost(), req.GetRestPrefix(), req.GetApplicationId())
	if err != nil {
		return nil, err
	}
	return toProtoService(svc), nil
}

// DeregisterService removes a registered service.
func (h *Handler) DeregisterService(ctx context.Context, req *api.DeregisterServiceRequest) (*emptypb.Empty, error) {
	if err := h.registry.DeregisterService(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// GetService returns a registered service.
func (h *Handler) GetService(ctx context.Context, req *api.GetServiceRequest) (*api.Service, error) {
	svc, err := h.registry.GetService(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return toProtoService(svc), nil
}

// ListServices lists all registered services.
func (h *Handler) ListServices(ctx context.Context, req *api.ListServicesRequest) (*api.ServiceList, error) {
	services, err := h.registry.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.Service, 0, len(services))
	for _, svc := range services {
		out = append(out, toProtoService(svc))
	}
	return &api.ServiceList{Services: out}, nil
}

// RegisterMicroApp registers a micro-app manifest.
func (h *Handler) RegisterMicroApp(ctx context.Context, req *api.RegisterMicroAppRequest) (*api.MicroApp, error) {
	app, err := h.registry.RegisterMicroApp(ctx, req.GetServiceName(), toDomainMicroApp(req.GetMicroApp(), req.GetApplicationId()))
	if err != nil {
		return nil, err
	}
	return toProtoMicroApp(app), nil
}

// ListMicroApps lists all registered micro-app manifests.
func (h *Handler) ListMicroApps(ctx context.Context, req *api.ListMicroAppsRequest) (*api.MicroAppList, error) {
	apps, err := h.registry.ListMicroApps(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.MicroApp, 0, len(apps))
	for _, a := range apps {
		out = append(out, toProtoMicroApp(a))
	}
	return &api.MicroAppList{MicroApps: out}, nil
}

// UpdateMicroApp updates a micro-app manifest.
func (h *Handler) UpdateMicroApp(ctx context.Context, req *api.UpdateMicroAppRequest) (*api.MicroApp, error) {
	app, err := h.registry.UpdateMicroApp(ctx, req.GetName(), toDomainMicroApp(&api.MicroApp{
		Name:              req.GetName(),
		Route:             req.GetRoute(),
		BundleUrl:         req.GetBundleUrl(),
		MenuLabelKey:      req.GetMenuLabelKey(),
		RequirePermission: req.GetRequirePermission(),
		Upstream:          req.GetUpstream(),
	}, ""))
	if err != nil {
		return nil, err
	}
	return toProtoMicroApp(app), nil
}

// DeleteMicroApp removes a micro-app manifest.
func (h *Handler) DeleteMicroApp(ctx context.Context, req *api.DeleteMicroAppRequest) (*emptypb.Empty, error) {
	if err := h.registry.DeleteMicroApp(ctx, req.GetName()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// CreateMenu creates a new menu item.
func (h *Handler) CreateMenu(ctx context.Context, req *api.CreateMenuRequest) (*api.Menu, error) {
	menu, err := h.registry.CreateMenu(ctx, toDomainMenu(&api.Menu{
		LabelKey:          req.GetLabelKey(),
		Route:             req.GetRoute(),
		Icon:              req.GetIcon(),
		ParentId:          req.GetParentId(),
		SortOrder:         req.GetSortOrder(),
		MicroAppName:      req.GetMicroAppName(),
		RequirePermission: req.GetRequirePermission(),
		ApplicationId:     req.GetApplicationId(),
	}))
	if err != nil {
		return nil, err
	}
	return toProtoMenu(menu), nil
}

// ListMenus lists all menu items.
func (h *Handler) ListMenus(ctx context.Context, req *api.ListMenusRequest) (*api.MenuList, error) {
	menus, err := h.registry.ListMenus(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.Menu, 0, len(menus))
	for _, m := range menus {
		out = append(out, toProtoMenu(m))
	}
	return &api.MenuList{Menus: out}, nil
}

// UpdateMenu updates a menu item.
func (h *Handler) UpdateMenu(ctx context.Context, req *api.UpdateMenuRequest) (*api.Menu, error) {
	menu, err := h.registry.UpdateMenu(ctx, toDomainMenu(&api.Menu{
		Id:                req.GetId(),
		LabelKey:          req.GetLabelKey(),
		Route:             req.GetRoute(),
		Icon:              req.GetIcon(),
		ParentId:          req.GetParentId(),
		SortOrder:         req.GetSortOrder(),
		MicroAppName:      req.GetMicroAppName(),
		RequirePermission: req.GetRequirePermission(),
		ApplicationId:     req.GetApplicationId(),
	}))
	if err != nil {
		return nil, err
	}
	return toProtoMenu(menu), nil
}

// DeleteMenu removes a menu item.
func (h *Handler) DeleteMenu(ctx context.Context, req *api.DeleteMenuRequest) (*emptypb.Empty, error) {
	if err := h.registry.DeleteMenu(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ReorderMenus updates sort order for menu items.
func (h *Handler) ReorderMenus(ctx context.Context, req *api.ReorderMenusRequest) (*api.MenuList, error) {
	order := make(map[string]int32, len(req.GetItems()))
	for _, item := range req.GetItems() {
		order[item.GetId()] = item.GetSortOrder()
	}
	menus, err := h.registry.ReorderMenus(ctx, order)
	if err != nil {
		return nil, err
	}
	out := make([]*api.Menu, 0, len(menus))
	for _, m := range menus {
		out = append(out, toProtoMenu(m))
	}
	return &api.MenuList{Menus: out}, nil
}

// SyncRoutes returns the gateway route manifest.
func (h *Handler) SyncRoutes(ctx context.Context, _ *api.SyncRoutesRequest) (*api.SyncRoutesResponse, error) {
	routes, err := h.registry.SyncRoutes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.ServiceRoute, 0, len(routes))
	for _, r := range routes {
		out = append(out, toProtoServiceRoute(r))
	}
	return &api.SyncRoutesResponse{Routes: out}, nil
}

// GetRoutePolicy returns the route policy for a service.
func (h *Handler) GetRoutePolicy(ctx context.Context, req *api.GetRoutePolicyRequest) (*api.RoutePolicy, error) {
	policy, err := h.registry.GetRoutePolicy(ctx, req.GetServiceId())
	if err != nil {
		return nil, err
	}
	return toProtoRoutePolicy(policy), nil
}

// SetRoutePolicy updates the route policy for a service.
func (h *Handler) SetRoutePolicy(ctx context.Context, req *api.SetRoutePolicyRequest) (*api.RoutePolicy, error) {
	policy, err := h.registry.SetRoutePolicy(ctx, req.GetServiceId(), toDomainRoutePolicy(req.GetPolicy()))
	if err != nil {
		return nil, err
	}
	return toProtoRoutePolicy(policy), nil
}

func toProtoApplication(app *domain.Application) *api.Application {
	if app == nil {
		return nil
	}
	return &api.Application{
		Id:          app.ID,
		Key:         app.Key,
		Name:        app.Name,
		LabelKey:    app.LabelKey,
		Icon:        app.Icon,
		Description: app.Description,
		Status:      toProtoApplicationStatus(app.Status),
		SortOrder:   app.SortOrder,
	}
}

func toDomainApplicationFromRegister(req *api.RegisterApplicationRequest) *domain.Application {
	if req == nil {
		return nil
	}
	return &domain.Application{
		Key:         req.GetKey(),
		Name:        req.GetName(),
		LabelKey:    req.GetLabelKey(),
		Icon:        req.GetIcon(),
		Description: req.GetDescription(),
		Status:      toDomainApplicationStatus(req.GetStatus()),
		SortOrder:   req.GetSortOrder(),
	}
}

func toDomainApplicationFromUpdate(req *api.UpdateApplicationRequest) *domain.Application {
	if req == nil {
		return nil
	}
	return &domain.Application{
		ID:          req.GetId(),
		Key:         req.GetKey(),
		Name:        req.GetName(),
		LabelKey:    req.GetLabelKey(),
		Icon:        req.GetIcon(),
		Description: req.GetDescription(),
		Status:      toDomainApplicationStatus(req.GetStatus()),
		SortOrder:   req.GetSortOrder(),
	}
}

func toProtoApplicationStatus(s domain.ApplicationStatus) api.ApplicationStatus {
	switch s {
	case domain.ApplicationStatusActive:
		return api.ApplicationStatus_APPLICATION_STATUS_ACTIVE
	case domain.ApplicationStatusOffline:
		return api.ApplicationStatus_APPLICATION_STATUS_OFFLINE
	default:
		return api.ApplicationStatus_APPLICATION_STATUS_UNSPECIFIED
	}
}

func toDomainApplicationStatus(s api.ApplicationStatus) domain.ApplicationStatus {
	switch s {
	case api.ApplicationStatus_APPLICATION_STATUS_ACTIVE:
		return domain.ApplicationStatusActive
	case api.ApplicationStatus_APPLICATION_STATUS_OFFLINE:
		return domain.ApplicationStatusOffline
	default:
		return domain.ApplicationStatusActive
	}
}

func toProtoResourceStatus(s domain.ResourceStatus) api.ResourceStatus {
	switch s {
	case domain.ResourceStatusDraft:
		return api.ResourceStatus_RESOURCE_STATUS_DRAFT
	case domain.ResourceStatusPending:
		return api.ResourceStatus_RESOURCE_STATUS_PENDING
	case domain.ResourceStatusOnline:
		return api.ResourceStatus_RESOURCE_STATUS_ONLINE
	case domain.ResourceStatusOffline:
		return api.ResourceStatus_RESOURCE_STATUS_OFFLINE
	case domain.ResourceStatusUpdating:
		return api.ResourceStatus_RESOURCE_STATUS_UPDATING
	default:
		return api.ResourceStatus_RESOURCE_STATUS_UNSPECIFIED
	}
}

func toDomainResourceStatus(s api.ResourceStatus) domain.ResourceStatus {
	switch s {
	case api.ResourceStatus_RESOURCE_STATUS_DRAFT:
		return domain.ResourceStatusDraft
	case api.ResourceStatus_RESOURCE_STATUS_PENDING:
		return domain.ResourceStatusPending
	case api.ResourceStatus_RESOURCE_STATUS_ONLINE:
		return domain.ResourceStatusOnline
	case api.ResourceStatus_RESOURCE_STATUS_OFFLINE:
		return domain.ResourceStatusOffline
	case api.ResourceStatus_RESOURCE_STATUS_UPDATING:
		return domain.ResourceStatusUpdating
	default:
		return domain.ResourceStatusOnline
	}
}

func toProtoService(svc *domain.Service) *api.Service {
	if svc == nil {
		return nil
	}
	routes := make([]*api.Route, 0, len(svc.Routes))
	for _, r := range svc.Routes {
		routes = append(routes, &api.Route{Path: r.Path, Method: r.Method})
	}
	microApps := make([]*api.MicroApp, 0, len(svc.MicroApps))
	for _, m := range svc.MicroApps {
		microApps = append(microApps, toProtoMicroApp(m))
	}
	return &api.Service{
		Id:             svc.ID,
		Name:           svc.Name,
		GrpcHost:       svc.GrpcHost,
		RestPrefix:     svc.RestPrefix,
		Routes:         routes,
		MicroApps:      microApps,
		ApplicationId:  svc.ApplicationID,
		ApplicationKey: svc.ApplicationKey,
		Status:         toProtoResourceStatus(svc.Status),
	}
}

func toDomainMicroApp(m *api.MicroApp, applicationID string) *domain.MicroApp {
	if m == nil {
		return nil
	}
	appID := applicationID
	if appID == "" {
		appID = m.GetApplicationId()
	}
	return &domain.MicroApp{
		Name:              m.GetName(),
		Route:             m.GetRoute(),
		BundleURL:         m.GetBundleUrl(),
		MenuLabelKey:      m.GetMenuLabelKey(),
		RequirePermission: m.GetRequirePermission(),
		ApplicationID:     appID,
		Upstream:          m.GetUpstream(),
		Status:            toDomainResourceStatus(m.GetStatus()),
	}
}

func toProtoMicroApp(m *domain.MicroApp) *api.MicroApp {
	if m == nil {
		return nil
	}
	return &api.MicroApp{
		Name:              m.Name,
		Route:             m.Route,
		BundleUrl:         m.BundleURL,
		MenuLabelKey:      m.MenuLabelKey,
		RequirePermission: m.RequirePermission,
		ApplicationId:     m.ApplicationID,
		ApplicationKey:    m.ApplicationKey,
		Upstream:          m.Upstream,
		Status:            toProtoResourceStatus(m.Status),
	}
}

func toDomainMenu(m *api.Menu) *domain.Menu {
	if m == nil {
		return nil
	}
	return &domain.Menu{
		ID:                m.GetId(),
		LabelKey:          m.GetLabelKey(),
		Route:             m.GetRoute(),
		Icon:              m.GetIcon(),
		ParentID:          m.GetParentId(),
		SortOrder:         m.GetSortOrder(),
		MicroAppName:      m.GetMicroAppName(),
		RequirePermission: m.GetRequirePermission(),
		ApplicationID:     m.GetApplicationId(),
		Status:            toDomainResourceStatus(m.GetStatus()),
	}
}

func toProtoMenu(m *domain.Menu) *api.Menu {
	if m == nil {
		return nil
	}
	return &api.Menu{
		Id:                m.ID,
		LabelKey:          m.LabelKey,
		Route:             m.Route,
		Icon:              m.Icon,
		ParentId:          m.ParentID,
		SortOrder:         m.SortOrder,
		MicroAppName:      m.MicroAppName,
		RequirePermission: m.RequirePermission,
		ApplicationId:     m.ApplicationID,
		ApplicationKey:    m.ApplicationKey,
		Status:            toProtoResourceStatus(m.Status),
	}
}

func toProtoRoutePolicy(p *domain.RoutePolicy) *api.RoutePolicy {
	if p == nil {
		return &api.RoutePolicy{AuthRequired: true}
	}
	return &api.RoutePolicy{
		RateLimitRps: p.RateLimitRPS,
		AuthRequired: p.AuthRequired,
		CanaryWeight: p.CanaryWeight,
		CanaryHost:   p.CanaryHost,
	}
}

func toDomainRoutePolicy(p *api.RoutePolicy) *domain.RoutePolicy {
	if p == nil {
		return &domain.RoutePolicy{AuthRequired: true}
	}
	return &domain.RoutePolicy{
		RateLimitRPS: p.GetRateLimitRps(),
		AuthRequired: p.GetAuthRequired(),
		CanaryWeight: p.GetCanaryWeight(),
		CanaryHost:   p.GetCanaryHost(),
	}
}

func toProtoServiceRoute(r *domain.ServiceRoute) *api.ServiceRoute {
	if r == nil {
		return nil
	}
	routes := make([]*api.Route, 0, len(r.Routes))
	for _, route := range r.Routes {
		if route == nil {
			continue
		}
		routes = append(routes, &api.Route{Path: route.Path, Method: route.Method})
	}
	return &api.ServiceRoute{
		ServiceId:    r.ServiceID,
		Name:         r.Name,
		RestPrefix:   r.RestPrefix,
		UpstreamHost: r.UpstreamHost,
		Routes:       routes,
		Policy:       toProtoRoutePolicy(r.Policy),
	}
}
