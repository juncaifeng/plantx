package grpc

import (
	"context"

	"github.com/plantx/platform/gateway-service/api"
	"github.com/plantx/platform/gateway-service/internal/app"
	"github.com/plantx/platform/gateway-service/internal/domain"
)

// Handler implements the GatewayService gRPC server.
type Handler struct {
	api.UnimplementedGatewayServiceServer
	registry *app.Registry
}

// NewHandler creates a new Handler.
func NewHandler(registry *app.Registry) *Handler {
	return &Handler{registry: registry}
}

// RegisterService registers a backend service.
func (h *Handler) RegisterService(ctx context.Context, req *api.RegisterServiceRequest) (*api.Service, error) {
	svc, err := h.registry.RegisterService(ctx, req.GetName(), req.GetGrpcHost(), req.GetRestPrefix())
	if err != nil {
		return nil, err
	}
	return toProtoService(svc), nil
}

// ListServices lists registered services.
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

// ListRoutes lists routes for a service.
func (h *Handler) ListRoutes(ctx context.Context, req *api.ListRoutesRequest) (*api.RouteList, error) {
	routes, ok, err := h.registry.GetRoutes(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	out := make([]*api.Route, 0, len(routes))
	for _, r := range routes {
		out = append(out, &api.Route{Path: r.Path, Method: r.Method})
	}
	return &api.RouteList{Routes: out}, nil
}

// RegisterMicroApp registers a micro-app manifest.
func (h *Handler) RegisterMicroApp(ctx context.Context, req *api.RegisterMicroAppRequest) (*api.MicroApp, error) {
	app, err := h.registry.RegisterMicroApp(ctx, req.GetServiceName(), toDomainMicroApp(req.GetMicroApp()))
	if err != nil {
		return nil, err
	}
	return toProtoMicroApp(app), nil
}

// ListMicroApps lists registered micro-app manifests.
func (h *Handler) ListMicroApps(ctx context.Context, req *api.ListMicroAppsRequest) (*api.MicroAppList, error) {
	apps, err := h.registry.ListMicroApps(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.MicroApp, 0, len(apps))
	for _, app := range apps {
		out = append(out, toProtoMicroApp(app))
	}
	return &api.MicroAppList{MicroApps: out}, nil
}

func toProtoService(svc *domain.Service) *api.Service {
	if svc == nil {
		return nil
	}
	routes := make([]*api.Route, 0, len(svc.Routes))
	for _, r := range svc.Routes {
		routes = append(routes, &api.Route{Path: r.Path, Method: r.Method})
	}
	return &api.Service{
		Id:         svc.ID,
		Name:       svc.Name,
		GrpcHost:   svc.GrpcHost,
		RestPrefix: svc.RestPrefix,
		Routes:     routes,
		MicroApp:   toProtoMicroApp(svc.MicroApp),
	}
}

func toDomainMicroApp(m *api.MicroApp) *domain.MicroApp {
	if m == nil {
		return nil
	}
	return &domain.MicroApp{
		Name:              m.GetName(),
		Route:             m.GetRoute(),
		BundleURL:         m.GetBundleUrl(),
		MenuLabelKey:      m.GetMenuLabelKey(),
		RequirePermission: m.GetRequirePermission(),
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
	}
}
