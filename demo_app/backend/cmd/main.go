package main

import (
	"context"
	"fmt"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	demoapi "github.com/plantx/demo_app/backend/api"
	"github.com/plantx/demo_app/backend/internal/app"
	"github.com/plantx/demo_app/backend/internal/infra/repo"
	grpcsrv "github.com/plantx/demo_app/backend/internal/interfaces/grpc"
	"github.com/plantx/kit/kit-go/auth"
	authmaxkey "github.com/plantx/kit/kit-go/auth/maxkey"
	"github.com/plantx/kit/kit-go/authz"
	authzopa "github.com/plantx/kit/kit-go/authz/opa"
	"github.com/plantx/kit/kit-go/config/env"
	"github.com/plantx/kit/kit-go/gateway"
	"github.com/plantx/kit/kit-go/log"
	"github.com/plantx/kit/kit-go/server"
	"github.com/plantx/kit/kit-go/tenant"
	registryapi "github.com/plantx/platform/registry-service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	stdLogger := stdlog.New(os.Stderr, "demo-service: ", stdlog.LstdFlags)

	cfg := env.New("DEMO")
	grpcPort := cfg.GetInt("grpc_port")
	if grpcPort == 0 {
		grpcPort = 8080
	}
	httpPort := cfg.GetInt("http_port")
	if httpPort == 0 {
		httpPort = 8081
	}

	// Use the shared in-memory repository. In production this would be postgres.
	repository := repo.NewInMemoryRepo()
	demoApp := app.NewDemoService(repository)

	// Authentication is optional: when MAXKEY_ISSUER/JWKS_URL/PUBLIC_KEY_PEM are
	// configured, kit-go validates JWTs; otherwise all requests are allowed.
	var authenticator auth.Authenticator
	maxkeyOpts := authmaxkey.EnvOptions("MAXKEY")
	if maxkeyOpts.Issuer != "" || maxkeyOpts.JWKSURL != "" || maxkeyOpts.PublicKeyPEM != "" {
		authenticator = authmaxkey.New(maxkeyOpts)
		stdLogger.Println("maxkey authentication enabled")
	}

	// Authorization is optional: when OPA_URL is configured, kit-go enforces the
	// plantx.kit.authz.action annotations defined in the proto file.
	var authorizer authz.Authorizer
	opaCfg := env.New("OPA")
	authorizer = authzopa.New(authzopa.Options{
		URL:          opaCfg.GetString("url"),
		DecisionPath: opaCfg.GetString("decision_path"),
	})

	registrar := gateway.AutoRegister("demo-service",
		gateway.WithApplication(gateway.Application{
			Key:       "demo",
			Name:      "Demo",
			LabelKey:  "nav.demo",
			SortOrder: 100,
			Status:    registryapi.ApplicationStatus_APPLICATION_STATUS_ACTIVE,
		}),
		gateway.WithMicroApp(gateway.MicroApp{
			Name:              "demo-ui",
			Route:             "/demo",
			BundleURL:         "/apps/demo-ui/demo-ui.js",
			MenuLabelKey:      "nav.demo",
			RequirePermission: "item:list",
			Upstream:          "demo-ui:80",
		}),
	)

	srv := server.New(server.Options{
		ServiceName:      "demo-service",
		GRPCPort:         grpcPort,
		HTTPPort:         httpPort,
		Logger:           log.FromContext(context.Background()),
		Authenticator:    authenticator,
		Authorizer:       authorizer,
		TenantResolver:   tenant.NewResolver(),
		GatewayRegistrar: registrar,
		Readiness: func(ctx context.Context) error {
			return nil
		},
	})

	handler := grpcsrv.NewHandler(demoApp)
	demoapi.RegisterDemoServiceServer(srv.GRPC(), handler)
	reflection.Register(srv.GRPC())

	if err := srv.RegisterGateway(context.Background(), demoapi.RegisterDemoServiceHandler); err != nil {
		stdLogger.Fatalf("failed to register gateway: %v", err)
	}
	stdLogger.Println("demo gateway registered")

	// Seed demo menus in registry-service after auto-registration has created
	// the application. kit-go/gateway.AutoRegister handles service/application/
	// micro-app registration; menu entities are created by calling the platform
	// registry-service API directly. This is a one-time bootstrap step, not a
	// kit-layer reimplementation.
	// Seed demo menus asynchronously: AutoRegister creates the application
	// during service startup, which may not have completed by the time we
	// first query the registry.
	go func() {
		if err := seedDemoMenus(stdLogger); err != nil {
			stdLogger.Printf("failed to seed demo menus: %v", err)
		}
	}()

	stdLogger.Printf("demo service starting on grpc_port=%d http_port=%d", grpcPort, httpPort)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := srv.Run(ctx); err != nil {
		stdLogger.Fatalf("server failed: %v", err)
	}
}

func seedDemoMenus(stdLogger *stdlog.Logger) error {
	registryAddr := os.Getenv("REGISTRY_SERVICE_GRPC_ADDR")
	if registryAddr == "" {
		registryAddr = "registry-service:8080"
	}

	// Retry a few times in case registry-service is still starting.
	var conn *grpc.ClientConn
	var err error
	for i := 0; i < 30; i++ {
		conn, err = grpc.NewClient(registryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	client := registryapi.NewRegistryServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var appID string
	for i := 0; i < 30; i++ {
		apps, err := client.ListApplications(ctx, &registryapi.ListApplicationsRequest{})
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		for _, a := range apps.GetApplications() {
			if a.GetKey() == "demo" {
				appID = a.GetId()
				break
			}
		}
		if appID != "" {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if appID == "" {
		return nil // Application not found; menu seeding skipped.
	}

	menus := []struct {
		labelKey string
		route    string
		icon     string
		perm     string
	}{
		{labelKey: "nav.demo.home", route: "/demo", icon: "HomeOutlined", perm: "item:list"},
		{labelKey: "nav.demo.config", route: "/demo/config", icon: "SettingOutlined", perm: "setting:list"},
		{labelKey: "nav.demo.system", route: "/demo/system", icon: "ToolOutlined", perm: "setting:admin"},
	}

	// Build a set of existing menus to avoid duplicates across restarts.
	existingResp, err := client.ListMenus(ctx, &registryapi.ListMenusRequest{})
	if err != nil {
		return fmt.Errorf("list menus: %w", err)
	}
	existing := make(map[string]struct{})
	for _, menu := range existingResp.GetMenus() {
		if menu.GetApplicationId() != appID {
			continue
		}
		key := menu.GetLabelKey() + "|" + menu.GetRoute()
		existing[key] = struct{}{}
	}

	for _, m := range menus {
		key := m.labelKey + "|" + m.route
		if _, ok := existing[key]; ok {
			stdLogger.Printf("menu %s already exists, skipping", m.labelKey)
			continue
		}
		_, err := client.CreateMenu(ctx, &registryapi.CreateMenuRequest{
			LabelKey:          m.labelKey,
			Route:             m.route,
			Icon:              m.icon,
			SortOrder:         10,
			MicroAppName:      "demo-ui",
			ApplicationId:     appID,
			RequirePermission: m.perm,
		})
		if err != nil {
			stdLogger.Printf("failed to create menu %s: %v", m.labelKey, err)
		} else {
			stdLogger.Printf("created menu %s", m.labelKey)
		}
	}
	return nil
}
