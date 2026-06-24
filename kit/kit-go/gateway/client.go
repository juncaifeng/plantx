// Package gateway provides a gRPC client for the platform registry service.
package gateway

import (
	"context"
	"fmt"

	"github.com/plantx/platform/registry-service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a gRPC client for the registry-service.
type Client struct {
	conn   *grpc.ClientConn
	client api.RegistryServiceClient
}

// NewClient creates a client connected to the registry-service at addr.
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

// RegisterApplication registers a new application with the registry.
func (c *Client) RegisterApplication(ctx context.Context, app *api.Application) (*api.Application, error) {
	return c.client.RegisterApplication(ctx, &api.RegisterApplicationRequest{
		Key:         app.GetKey(),
		Name:        app.GetName(),
		LabelKey:    app.GetLabelKey(),
		Icon:        app.GetIcon(),
		Description: app.GetDescription(),
		Status:      app.GetStatus(),
		SortOrder:   app.GetSortOrder(),
	})
}

// GetApplication returns an application by id.
func (c *Client) GetApplication(ctx context.Context, id string) (*api.Application, error) {
	return c.client.GetApplication(ctx, &api.GetApplicationRequest{Id: id})
}

// ListApplications returns the registered applications.
func (c *Client) ListApplications(ctx context.Context) ([]*api.Application, error) {
	resp, err := c.client.ListApplications(ctx, &api.ListApplicationsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetApplications(), nil
}

// RegisterService registers a backend service with the registry.
func (c *Client) RegisterService(ctx context.Context, name, grpcHost, restPrefix, applicationID, applicationKey string) (*api.Service, error) {
	return c.client.RegisterService(ctx, &api.RegisterServiceRequest{
		Name:           name,
		GrpcHost:       grpcHost,
		RestPrefix:     restPrefix,
		ApplicationId:  applicationID,
		ApplicationKey: applicationKey,
	})
}

// DeregisterService removes a registered service.
func (c *Client) DeregisterService(ctx context.Context, id string) error {
	_, err := c.client.DeregisterService(ctx, &api.DeregisterServiceRequest{Id: id})
	return err
}

// RegisterMicroApp registers a micro-app manifest for a service.
func (c *Client) RegisterMicroApp(ctx context.Context, serviceName string, microApp *api.MicroApp, applicationID, applicationKey string) (*api.MicroApp, error) {
	return c.client.RegisterMicroApp(ctx, &api.RegisterMicroAppRequest{
		ServiceName:    serviceName,
		MicroApp:       microApp,
		ApplicationId:  applicationID,
		ApplicationKey: applicationKey,
	})
}

// ListMicroApps returns the registered micro-apps.
func (c *Client) ListMicroApps(ctx context.Context) ([]*api.MicroApp, error) {
	resp, err := c.client.ListMicroApps(ctx, &api.ListMicroAppsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetMicroApps(), nil
}

// RegisterMenu registers a menu item with the registry.
func (c *Client) RegisterMenu(ctx context.Context, menu *api.Menu, applicationID, applicationKey string) (*api.Menu, error) {
	return c.client.CreateMenu(ctx, &api.CreateMenuRequest{
		LabelKey:          menu.GetLabelKey(),
		Route:             menu.GetRoute(),
		Icon:              menu.GetIcon(),
		ParentId:          menu.GetParentId(),
		SortOrder:         menu.GetSortOrder(),
		MicroAppName:      menu.GetMicroAppName(),
		RequirePermission: menu.GetRequirePermission(),
		ApplicationId:     applicationID,
		ApplicationKey:    applicationKey,
		Status:            menu.GetStatus(),
	})
}

// Close closes the underlying gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
