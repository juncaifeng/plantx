package gateway

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/plantx/platform/registry-service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// testRegistryServer is a minimal in-memory registry service for tests.
type testRegistryServer struct {
	api.UnimplementedRegistryServiceServer
	mu           sync.RWMutex
	applications map[string]*api.Application
	services     map[string]*api.Service
	microApps    map[string]*api.MicroApp
	menus        map[string]*api.Menu
}

func newTestRegistryServer() *testRegistryServer {
	return &testRegistryServer{
		applications: make(map[string]*api.Application),
		services:     make(map[string]*api.Service),
		microApps:    make(map[string]*api.MicroApp),
		menus:        make(map[string]*api.Menu),
	}
}

func (s *testRegistryServer) RegisterApplication(_ context.Context, req *api.RegisterApplicationRequest) (*api.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.applications {
		if existing.GetKey() == req.GetKey() {
			return nil, fmt.Errorf("application with key %q already exists", req.GetKey())
		}
	}

	app := &api.Application{
		Id:          uuid.NewString(),
		Key:         req.GetKey(),
		Name:        req.GetName(),
		LabelKey:    req.GetLabelKey(),
		Icon:        req.GetIcon(),
		Description: req.GetDescription(),
		Status:      req.GetStatus(),
		SortOrder:   req.GetSortOrder(),
	}
	s.applications[app.GetId()] = app
	return app, nil
}

func (s *testRegistryServer) ListApplications(_ context.Context, _ *api.ListApplicationsRequest) (*api.ApplicationList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*api.Application, 0, len(s.applications))
	for _, app := range s.applications {
		out = append(out, app)
	}
	return &api.ApplicationList{Applications: out}, nil
}

func (s *testRegistryServer) RegisterService(_ context.Context, req *api.RegisterServiceRequest) (*api.Service, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	svc := &api.Service{
		Id:             uuid.NewString(),
		Name:           req.GetName(),
		GrpcHost:       req.GetGrpcHost(),
		RestPrefix:     req.GetRestPrefix(),
		ApplicationId:  req.GetApplicationId(),
		ApplicationKey: req.GetApplicationKey(),
	}
	s.services[svc.GetName()] = svc
	return svc, nil
}

func (s *testRegistryServer) DeregisterService(_ context.Context, _ *api.DeregisterServiceRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *testRegistryServer) GetService(_ context.Context, _ *api.GetServiceRequest) (*api.Service, error) {
	return &api.Service{}, nil
}

func (s *testRegistryServer) ListServices(_ context.Context, _ *api.ListServicesRequest) (*api.ServiceList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*api.Service, 0, len(s.services))
	for _, svc := range s.services {
		out = append(out, svc)
	}
	return &api.ServiceList{Services: out}, nil
}

func (s *testRegistryServer) RegisterMicroApp(_ context.Context, req *api.RegisterMicroAppRequest) (*api.MicroApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m := req.GetMicroApp()
	if m == nil {
		return nil, fmt.Errorf("micro-app is required")
	}
	m.ApplicationId = req.GetApplicationId()
	m.ApplicationKey = req.GetApplicationKey()
	s.microApps[m.GetName()] = m
	return m, nil
}

func (s *testRegistryServer) ListMicroApps(_ context.Context, _ *api.ListMicroAppsRequest) (*api.MicroAppList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*api.MicroApp, 0, len(s.microApps))
	for _, m := range s.microApps {
		out = append(out, m)
	}
	return &api.MicroAppList{MicroApps: out}, nil
}

func (s *testRegistryServer) CreateMenu(_ context.Context, req *api.CreateMenuRequest) (*api.Menu, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := &api.Menu{
		Id:                fmt.Sprintf("menu-%d", len(s.menus)+1),
		LabelKey:          req.GetLabelKey(),
		Route:             req.GetRoute(),
		Icon:              req.GetIcon(),
		ParentId:          req.GetParentId(),
		SortOrder:         req.GetSortOrder(),
		MicroAppName:      req.GetMicroAppName(),
		RequirePermission: req.GetRequirePermission(),
		ApplicationId:     req.GetApplicationId(),
		ApplicationKey:    req.GetApplicationKey(),
		Status:            req.GetStatus(),
	}
	s.menus[m.GetId()] = m
	return m, nil
}

func startTestRegistry(t *testing.T) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	api.RegisterRegistryServiceServer(grpcServer, newTestRegistryServer())
	go func() { _ = grpcServer.Serve(lis) }()
	return lis.Addr().String(), func() { grpcServer.Stop() }
}

func TestAutoRegister(t *testing.T) {
	addr, cleanup := startTestRegistry(t)
	defer cleanup()

	reg := AutoRegister("order-service",
		WithRegistryAddr(addr),
		WithMicroApp(MicroApp{
			Name:         "order-ui",
			Route:        "/order",
			BundleURL:    "/apps/order-ui/order-ui.js",
			MenuLabelKey: "nav.order",
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := reg.Register(ctx); err != nil {
		t.Fatalf("register: %v", err)
	}
	defer func() { _ = reg.Deregister(ctx) }()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	client := api.NewRegistryServiceClient(conn)

	svcs, err := client.ListServices(ctx, &api.ListServicesRequest{})
	if err != nil {
		t.Fatalf("list services: %v", err)
	}
	found := false
	for _, s := range svcs.GetServices() {
		if s.GetName() == "order-service" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("order-service not registered: %v", svcs.GetServices())
	}

	apps, err := client.ListMicroApps(ctx, &api.ListMicroAppsRequest{})
	if err != nil {
		t.Fatalf("list micro apps: %v", err)
	}
	foundApp := false
	for _, a := range apps.GetMicroApps() {
		if a.GetName() == "order-ui" && a.GetRoute() == "/order" {
			foundApp = true
			break
		}
	}
	if !foundApp {
		t.Fatalf("order-ui micro-app not registered: %v", apps.GetMicroApps())
	}
}

func TestAutoRegisterWithApplication(t *testing.T) {
	addr, cleanup := startTestRegistry(t)
	defer cleanup()

	reg := AutoRegister("order-service",
		WithRegistryAddr(addr),
		WithApplication(Application{
			Key:      "order",
			Name:     "Order",
			LabelKey: "nav.orders",
		}),
		WithMicroApp(MicroApp{
			Name:         "order-ui",
			Route:        "/order",
			BundleURL:    "/apps/order-ui/order-ui.js",
			MenuLabelKey: "nav.order",
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := reg.Register(ctx); err != nil {
		t.Fatalf("register: %v", err)
	}
	defer func() { _ = reg.Deregister(ctx) }()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	client := api.NewRegistryServiceClient(conn)

	apps, err := client.ListApplications(ctx, &api.ListApplicationsRequest{})
	if err != nil {
		t.Fatalf("list applications: %v", err)
	}
	var orderApp *api.Application
	for _, a := range apps.GetApplications() {
		if a.GetKey() == "order" {
			orderApp = a
			break
		}
	}
	if orderApp == nil {
		t.Fatalf("order application not registered: %v", apps.GetApplications())
	}
	if orderApp.GetName() != "Order" || orderApp.GetLabelKey() != "nav.orders" {
		t.Fatalf("unexpected application fields: %v", orderApp)
	}

	svcs, err := client.ListServices(ctx, &api.ListServicesRequest{})
	if err != nil {
		t.Fatalf("list services: %v", err)
	}
	var orderSvc *api.Service
	for _, s := range svcs.GetServices() {
		if s.GetName() == "order-service" {
			orderSvc = s
			break
		}
	}
	if orderSvc == nil {
		t.Fatalf("order-service not registered")
	}
	if orderSvc.GetApplicationId() != orderApp.GetId() {
		t.Fatalf("service application_id mismatch: got %q, want %q", orderSvc.GetApplicationId(), orderApp.GetId())
	}

	micros, err := client.ListMicroApps(ctx, &api.ListMicroAppsRequest{})
	if err != nil {
		t.Fatalf("list micro apps: %v", err)
	}
	var orderUI *api.MicroApp
	for _, m := range micros.GetMicroApps() {
		if m.GetName() == "order-ui" {
			orderUI = m
			break
		}
	}
	if orderUI == nil {
		t.Fatalf("order-ui micro-app not registered")
	}
	if orderUI.GetApplicationId() != orderApp.GetId() {
		t.Fatalf("micro-app application_id mismatch: got %q, want %q", orderUI.GetApplicationId(), orderApp.GetId())
	}
}

func TestAutoRegisterWithSharedApplication(t *testing.T) {
	addr, cleanup := startTestRegistry(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app := Application{Key: "platform", Name: "Platform", LabelKey: "nav.platform"}

	reg1 := AutoRegister("iam-service",
		WithRegistryAddr(addr),
		WithApplication(app),
		WithMicroApp(MicroApp{Name: "iam-admin-ui", Route: "/admin/iam"}),
	)
	if err := reg1.Register(ctx); err != nil {
		t.Fatalf("register first service: %v", err)
	}
	defer func() { _ = reg1.Deregister(ctx) }()

	reg2 := AutoRegister("tenant-service",
		WithRegistryAddr(addr),
		WithApplication(app),
		WithMicroApp(MicroApp{Name: "tenant-admin-ui", Route: "/admin/tenant"}),
	)
	if err := reg2.Register(ctx); err != nil {
		t.Fatalf("register second service: %v", err)
	}
	defer func() { _ = reg2.Deregister(ctx) }()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	client := api.NewRegistryServiceClient(conn)

	apps, err := client.ListApplications(ctx, &api.ListApplicationsRequest{})
	if err != nil {
		t.Fatalf("list applications: %v", err)
	}
	platformCount := 0
	for _, a := range apps.GetApplications() {
		if a.GetKey() == "platform" {
			platformCount++
		}
	}
	if platformCount != 1 {
		t.Fatalf("expected exactly one platform application, got %d", platformCount)
	}
}
