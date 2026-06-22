package gateway

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/plantx/kit/kit-go/server"
	"github.com/plantx/platform/registry-service/api"
)

// MicroApp describes a qiankun micro-frontend manifest.
type MicroApp struct {
	Name              string
	Route             string
	BundleURL         string
	MenuLabelKey      string
	RequirePermission string
}

// Application describes a product / application that groups services, micro-apps and menus.
type Application struct {
	Key         string
	Name        string
	LabelKey    string
	Icon        string
	Description string
	Status      api.ApplicationStatus
	SortOrder   int32
}

func toPBMicroApp(m *MicroApp) *api.MicroApp {
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

type options struct {
	registryAddr string
	grpcHost     string
	restPrefix   string
	microApps    []*MicroApp
	application  *Application
}

// Option configures AutoRegister.
type Option func(*options)

// WithMicroApp attaches micro-frontend metadata to the service registration.
// May be called multiple times to register multiple micro-apps for one service.
func WithMicroApp(m MicroApp) Option {
	return func(o *options) {
		o.microApps = append(o.microApps, &m)
	}
}

// WithApplication attaches the service and its micro-apps to a product application.
// If the application key does not yet exist in the registry it is created; otherwise
// the existing application is reused.
func WithApplication(a Application) Option {
	return func(o *options) {
		o.application = &a
	}
}

// WithGRPCHost overrides the externally reachable gRPC host advertised to the registry.
func WithGRPCHost(host string) Option {
	return func(o *options) {
		o.grpcHost = host
	}
}

// WithRESTPrefix overrides the REST prefix advertised to the registry.
func WithRESTPrefix(prefix string) Option {
	return func(o *options) {
		o.restPrefix = prefix
	}
}

// WithRegistryAddr overrides the registry-service gRPC address.
func WithRegistryAddr(addr string) Option {
	return func(o *options) {
		o.registryAddr = addr
	}
}

// WithGatewayAddr is a deprecated alias for WithRegistryAddr.
func WithGatewayAddr(addr string) Option {
	return WithRegistryAddr(addr)
}

type registrar struct {
	serviceName    string
	opts           options
	serviceID      string
	applicationID  string
	applicationKey string
	client         *Client
}

// AutoRegister returns a server.GatewayRegistrar that registers the service
// with the registry-service when the server starts.
func AutoRegister(serviceName string, opts ...Option) server.GatewayRegistrar {
	o := options{
		registryAddr: defaultRegistryAddr(),
		grpcHost:     defaultGRPCHost(serviceName),
		restPrefix:   defaultRESTPrefix(serviceName),
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &registrar{serviceName: serviceName, opts: o}
}

func (r *registrar) Register(ctx context.Context) error {
	client, err := NewClient(r.opts.registryAddr)
	if err != nil {
		return err
	}
	r.client = client

	if r.opts.application != nil {
		app, err := r.ensureApplication(ctx, client, r.opts.application)
		if err != nil {
			_ = client.Close()
			return fmt.Errorf("ensure application %s: %w", r.opts.application.Key, err)
		}
		r.applicationID = app.GetId()
		r.applicationKey = app.GetKey()
	}

	svc, err := client.RegisterService(ctx, r.serviceName, r.opts.grpcHost, r.opts.restPrefix, r.applicationID, r.applicationKey)
	if err != nil {
		_ = client.Close()
		return fmt.Errorf("register service %s: %w", r.serviceName, err)
	}
	r.serviceID = svc.GetId()

	for _, m := range r.opts.microApps {
		if _, err := client.RegisterMicroApp(ctx, r.serviceName, toPBMicroApp(m), r.applicationID, r.applicationKey); err != nil {
			_ = client.Close()
			return fmt.Errorf("register micro-app %s: %w", r.serviceName, err)
		}
	}
	return nil
}

func (r *registrar) ensureApplication(ctx context.Context, client *Client, app *Application) (*api.Application, error) {
	if existing, err := r.findApplicationByKey(ctx, client, app.Key); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	status := app.Status
	if status == api.ApplicationStatus_APPLICATION_STATUS_UNSPECIFIED {
		status = api.ApplicationStatus_APPLICATION_STATUS_ACTIVE
	}

	created, err := client.RegisterApplication(ctx, &api.Application{
		Key:         app.Key,
		Name:        app.Name,
		LabelKey:    app.LabelKey,
		Icon:        app.Icon,
		Description: app.Description,
		Status:      status,
		SortOrder:   app.SortOrder,
	})
	if err == nil {
		return created, nil
	}

	// A concurrent registrar may have created the application; fall back to a lookup.
	if existing, findErr := r.findApplicationByKey(ctx, client, app.Key); findErr != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}
	return nil, err
}

func (r *registrar) findApplicationByKey(ctx context.Context, client *Client, key string) (*api.Application, error) {
	apps, err := client.ListApplications(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range apps {
		if a.GetKey() == key {
			return a, nil
		}
	}
	return nil, nil
}

func (r *registrar) Deregister(ctx context.Context) error {
	if r.client == nil {
		return nil
	}
	_ = r.client.DeregisterService(ctx, r.serviceID)
	return r.client.Close()
}

func defaultRegistryAddr() string {
	if v := os.Getenv("REGISTRY_SERVICE_GRPC_ADDR"); v != "" {
		return v
	}
	// Backward compatibility: allow legacy GATEWAY_SERVICE_GRPC_ADDR to be used
	// when the new registry env var is not set.
	if v := os.Getenv("GATEWAY_SERVICE_GRPC_ADDR"); v != "" {
		return v
	}
	return "registry-service:8080"
}

func defaultGRPCHost(serviceName string) string {
	key := strings.ToUpper(strings.ReplaceAll(serviceName, "-", "_")) + "_GRPC_HOST"
	if v := os.Getenv(key); v != "" {
		return v
	}
	return serviceName + ":8080"
}

func defaultRESTPrefix(serviceName string) string {
	key := strings.ToUpper(strings.ReplaceAll(serviceName, "-", "_")) + "_REST_PREFIX"
	if v := os.Getenv(key); v != "" {
		return v
	}
	base := strings.TrimSuffix(serviceName, "-service")
	return "/api/" + base + "/v1"
}
