// Package registry provides a registry-service client for the gateway.
package registry

import (
	"context"
	"fmt"

	"github.com/plantx/platform/gateway-service/internal/domain"
	"github.com/plantx/platform/registry-service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a registry-service client that implements domain.Repository.
type Client struct {
	conn   *grpc.ClientConn
	client api.RegistryServiceClient
}

// NewClient creates a client connected to registry-service at addr.
func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial registry-service %s: %w", addr, err)
	}
	return &Client{
		conn:   conn,
		client: api.NewRegistryServiceClient(conn),
	}, nil
}

// Register registers a service with the registry-service.
func (c *Client) Register(ctx context.Context, name, grpcHost, restPrefix string, microApp *domain.MicroApp) (*domain.Service, error) {
	svc, err := c.client.RegisterService(ctx, &api.RegisterServiceRequest{
		Name:       name,
		GrpcHost:   grpcHost,
		RestPrefix: restPrefix,
	})
	if err != nil {
		return nil, err
	}
	if microApp != nil {
		if _, err := c.client.RegisterMicroApp(ctx, &api.RegisterMicroAppRequest{
			ServiceName: name,
			MicroApp:    toPBMicroApp(microApp),
		}); err != nil {
			return nil, err
		}
	}
	return toDomainService(svc), nil
}

// List returns all registered services.
func (c *Client) List(ctx context.Context) ([]*domain.Service, error) {
	resp, err := c.client.ListServices(ctx, &api.ListServicesRequest{})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Service, 0, len(resp.GetServices()))
	for _, svc := range resp.GetServices() {
		out = append(out, toDomainService(svc))
	}
	return out, nil
}

// Routes returns the route derived from a service's REST prefix.
func (c *Client) Routes(ctx context.Context, id string) ([]*domain.Route, bool, error) {
	svc, err := c.client.GetService(ctx, &api.GetServiceRequest{Id: id})
	if err != nil {
		return nil, false, err
	}
	if svc == nil {
		return nil, false, nil
	}
	routes := make([]*domain.Route, 0, len(svc.GetRoutes()))
	for _, r := range svc.GetRoutes() {
		routes = append(routes, &domain.Route{Path: r.GetPath(), Method: r.GetMethod()})
	}
	if len(routes) == 0 {
		routes = append(routes, &domain.Route{Path: svc.GetRestPrefix(), Method: "*"})
	}
	return routes, true, nil
}

// RegisterMicroApp registers a micro-app manifest for a service.
func (c *Client) RegisterMicroApp(ctx context.Context, serviceName string, microApp *domain.MicroApp) (*domain.MicroApp, error) {
	app, err := c.client.RegisterMicroApp(ctx, &api.RegisterMicroAppRequest{
		ServiceName: serviceName,
		MicroApp:    toPBMicroApp(microApp),
	})
	if err != nil {
		return nil, err
	}
	return toDomainMicroApp(app), nil
}

// ListMicroApps returns all registered micro-app manifests.
func (c *Client) ListMicroApps(ctx context.Context) ([]*domain.MicroApp, error) {
	resp, err := c.client.ListMicroApps(ctx, &api.ListMicroAppsRequest{})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.MicroApp, 0, len(resp.GetMicroApps()))
	for _, app := range resp.GetMicroApps() {
		out = append(out, toDomainMicroApp(app))
	}
	return out, nil
}

// Close closes the underlying gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

func toDomainService(svc *api.Service) *domain.Service {
	if svc == nil {
		return nil
	}
	routes := make([]*domain.Route, 0, len(svc.GetRoutes()))
	for _, r := range svc.GetRoutes() {
		routes = append(routes, &domain.Route{Path: r.GetPath(), Method: r.GetMethod()})
	}
	var microApp *domain.MicroApp
	if microApps := svc.GetMicroApps(); len(microApps) > 0 {
		microApp = toDomainMicroApp(microApps[0])
	}
	return &domain.Service{
		ID:         svc.GetId(),
		Name:       svc.GetName(),
		GrpcHost:   svc.GetGrpcHost(),
		RestPrefix: svc.GetRestPrefix(),
		Routes:     routes,
		MicroApp:   microApp,
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

func toPBMicroApp(m *domain.MicroApp) *api.MicroApp {
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
