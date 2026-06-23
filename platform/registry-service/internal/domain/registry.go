package domain

import "context"

// ApplicationStatus represents the lifecycle status of an application.
type ApplicationStatus string

const (
	ApplicationStatusActive  ApplicationStatus = "ACTIVE"
	ApplicationStatusOffline ApplicationStatus = "OFFLINE"
)

// Application represents a product / application that groups services, micro-apps and menus.
type Application struct {
	ID          string
	Key         string
	Name        string
	LabelKey    string
	Icon        string
	Description string
	Status      ApplicationStatus
	SortOrder   int32
}

// Route represents a REST route exposed by a service.
type Route struct {
	Path   string
	Method string
}

// MicroApp represents a qiankun micro-frontend manifest.
type MicroApp struct {
	Name              string
	Route             string
	BundleURL         string
	MenuLabelKey      string
	RequirePermission string
	ApplicationID     string
	ApplicationKey    string
	Upstream          string
	Status            ResourceStatus
}

// RoutePolicy represents gateway routing/edge policies for a service.
type RoutePolicy struct {
	RateLimitRPS int32
	AuthRequired bool
	CanaryWeight int32
	CanaryHost   string
}

// ServiceRoute is the gateway-facing route manifest for a service.
type ServiceRoute struct {
	ServiceID    string
	Name         string
	RestPrefix   string
	UpstreamHost string
	Routes       []*Route
	Policy       *RoutePolicy
}

// Service represents a registered backend service.
type Service struct {
	ID             string
	Name           string
	GrpcHost       string
	RestPrefix     string
	Routes         []*Route
	MicroApps      []*MicroApp
	Policy         *RoutePolicy
	ApplicationID  string
	ApplicationKey string
	Status         ResourceStatus
}

// Menu represents a configurable portal menu item.
type Menu struct {
	ID                string
	LabelKey          string
	Route             string
	Icon              string
	ParentID          string
	SortOrder         int32
	MicroAppName      string
	RequirePermission string
	ApplicationID     string
	ApplicationKey    string
	Status            ResourceStatus
}

// Repository defines persistence operations for the registry.
type Repository interface {
	// Applications
	RegisterApplication(ctx context.Context, app *Application) (*Application, error)
	ListApplications(ctx context.Context) ([]*Application, error)
	GetApplication(ctx context.Context, id string) (*Application, error)
	UpdateApplication(ctx context.Context, app *Application) (*Application, error)
	DeleteApplication(ctx context.Context, id string) error
	GetApplicationMenus(ctx context.Context, applicationID string) ([]*Menu, error)
	GetApplicationMicroApps(ctx context.Context, applicationID string) ([]*MicroApp, error)

	// Services
	RegisterService(ctx context.Context, name, grpcHost, restPrefix, applicationID string) (*Service, error)
	DeregisterService(ctx context.Context, id string) error
	UpdateServiceStatus(ctx context.Context, name string, status ResourceStatus) (*Service, error)
	GetService(ctx context.Context, id string) (*Service, error)
	ListServices(ctx context.Context) ([]*Service, error)

	// MicroApps
	RegisterMicroApp(ctx context.Context, serviceName string, microApp *MicroApp) (*MicroApp, error)
	ListMicroApps(ctx context.Context) ([]*MicroApp, error)
	UpdateMicroApp(ctx context.Context, name string, microApp *MicroApp) (*MicroApp, error)
	DeleteMicroApp(ctx context.Context, name string) error

	// Menus
	CreateMenu(ctx context.Context, menu *Menu) (*Menu, error)
	ListMenus(ctx context.Context) ([]*Menu, error)
	UpdateMenu(ctx context.Context, menu *Menu) (*Menu, error)
	DeleteMenu(ctx context.Context, id string) error
	ReorderMenus(ctx context.Context, order map[string]int32) ([]*Menu, error)

	// Route policies
	GetRoutePolicy(ctx context.Context, serviceID string) (*RoutePolicy, error)
	SetRoutePolicy(ctx context.Context, serviceID string, policy *RoutePolicy) (*RoutePolicy, error)
	DeleteRoutePolicy(ctx context.Context, serviceID string) error
	ListRoutePolicies(ctx context.Context) (map[string]*RoutePolicy, error)
}
