package domain

import "context"

// Route represents a REST route exposed by a service.
type Route struct {
	Path   string
	Method string
}

// MicroApp represents a micro-frontend manifest.
type MicroApp struct {
	Name              string
	Route             string
	BundleURL         string
	MenuLabelKey      string
	RequirePermission string
}

// Service represents a registered backend service.
type Service struct {
	ID         string
	Name       string
	GrpcHost   string
	RestPrefix string
	Routes     []*Route
	MicroApp   *MicroApp
}

// Repository defines persistence operations for the gateway registry.
type Repository interface {
	Register(ctx context.Context, name, grpcHost, restPrefix string, microApp *MicroApp) (*Service, error)
	List(ctx context.Context) ([]*Service, error)
	Routes(ctx context.Context, id string) ([]*Route, bool, error)
	RegisterMicroApp(ctx context.Context, serviceName string, microApp *MicroApp) (*MicroApp, error)
	ListMicroApps(ctx context.Context) ([]*MicroApp, error)
}
