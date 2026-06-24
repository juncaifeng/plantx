package main

import (
	"context"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

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
		gateway.WithMenu(gateway.Menu{
			LabelKey:          "nav.demo.home",
			Route:             "/demo",
			Icon:              "HomeOutlined",
			SortOrder:         10,
			MicroAppName:      "demo-ui",
			RequirePermission: "item:list",
		}),
		gateway.WithMenu(gateway.Menu{
			LabelKey:          "nav.demo.config",
			Route:             "/demo/config",
			Icon:              "SettingOutlined",
			SortOrder:         20,
			MicroAppName:      "demo-ui",
			RequirePermission: "setting:list",
		}),
		gateway.WithMenu(gateway.Menu{
			LabelKey:          "nav.demo.system",
			Route:             "/demo/system",
			Icon:              "ToolOutlined",
			SortOrder:         30,
			MicroAppName:      "demo-ui",
			RequirePermission: "setting:admin",
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

	stdLogger.Printf("demo service starting on grpc_port=%d http_port=%d", grpcPort, httpPort)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := srv.Run(ctx); err != nil {
		stdLogger.Fatalf("server failed: %v", err)
	}
}
