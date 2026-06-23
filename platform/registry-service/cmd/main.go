package main

import (
	"context"
	"log"
	"os"

	"github.com/plantx/kit/kit-go/auth"
	authmaxkey "github.com/plantx/kit/kit-go/auth/maxkey"
	"github.com/plantx/kit/kit-go/authz"
	authzopa "github.com/plantx/kit/kit-go/authz/opa"
	"github.com/plantx/kit/kit-go/config/env"
	"github.com/plantx/kit/kit-go/db"
	"github.com/plantx/kit/kit-go/db/postgres"
	"github.com/plantx/kit/kit-go/gateway"
	kitlog "github.com/plantx/kit/kit-go/log"
	zaplog "github.com/plantx/kit/kit-go/log/zap"
	"github.com/plantx/kit/kit-go/server"
	"github.com/plantx/kit/kit-go/tenant"
	"github.com/plantx/platform/registry-service/api"
	"github.com/plantx/platform/registry-service/internal/app"
	"github.com/plantx/platform/registry-service/internal/infra/migrate"
	"github.com/plantx/platform/registry-service/internal/infra/repo"
	"github.com/plantx/platform/registry-service/internal/infra/temporal"
	grpcsrv "github.com/plantx/platform/registry-service/internal/interfaces/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	stdLogger := log.New(os.Stderr, "registry-service: ", log.LstdFlags)

	logger, err := zaplog.New()
	if err != nil {
		stdLogger.Printf("failed to create logger: %v", err)
		os.Exit(1)
	}

	cfg := env.New("REGISTRY")
	grpcPort := cfg.GetInt("grpc_port")
	if grpcPort == 0 {
		grpcPort = 8080
	}
	httpPort := cfg.GetInt("http_port")
	if httpPort == 0 {
		httpPort = 8081
	}

	var sqldb db.DB
	if dsn := cfg.GetString("database_dsn"); dsn != "" {
		sqldb, err = postgres.New(dsn)
		if err != nil {
			logger.Error("failed to connect to postgres", kitlog.F("error", err))
			os.Exit(1)
		}
	} else {
		logger.Error("DATABASE_DSN is required")
		os.Exit(1)
	}

	readiness := func(ctx context.Context) error {
		return sqldb.PingContext(ctx)
	}

	var authenticator auth.Authenticator
	maxkeyOpts := authmaxkey.EnvOptions("MAXKEY")
	if maxkeyOpts.Issuer != "" || maxkeyOpts.JWKSURL != "" || maxkeyOpts.PublicKeyPEM != "" {
		authenticator = authmaxkey.New(maxkeyOpts)
		logger.Info("maxkey authentication enabled")
	} else {
		logger.Warn("authentication disabled: configure MAXKEY_ISSUER, MAXKEY_JWKS_URL or MAXKEY_PUBLIC_KEY_PEM")
	}

	var authorizer authz.Authorizer
	opaCfg := env.New("OPA")
	authorizer = authzopa.New(authzopa.Options{
		URL:          opaCfg.GetString("url"),
		DecisionPath: opaCfg.GetString("decision_path"),
	})

	srv := server.New(server.Options{
		ServiceName:    "registry-service",
		GRPCPort:       grpcPort,
		HTTPPort:       httpPort,
		Authenticator:  authenticator,
		Authorizer:     authorizer,
		TenantResolver: tenant.NewResolver(),
		Logger:         logger,
		Config:         cfg,
		DB:             sqldb,
		Readiness:      readiness,
		GatewayRegistrar: gateway.AutoRegister("registry-service",
			gateway.WithApplication(gateway.Application{
				Key:      "platform",
				Name:     "Platform",
				LabelKey: "nav.admin",
			}),
			gateway.WithMicroApp(gateway.MicroApp{
				Name:              "registry-admin-ui",
				Route:             "/admin/registry",
				BundleURL:         "/apps/registry-admin-ui/registry-admin-ui.js",
				MenuLabelKey:      "nav.registry",
				RequirePermission: "registry:read",
			}),
			gateway.WithMicroApp(gateway.MicroApp{
				Name:              "api-explorer-ui",
				Route:             "/admin/api-explorer",
				BundleURL:         "/apps/api-explorer-ui/api-explorer-ui.js",
				MenuLabelKey:      "nav.apiExplorer",
				RequirePermission: "registry:read",
			}),
		),
		AuthExcludedMethods: []string{
			"/plantx.registry.v1.RegistryService/RegisterService",
			"/plantx.registry.v1.RegistryService/DeregisterService",
			"/plantx.registry.v1.RegistryService/UpdateServiceStatus",
			"/plantx.registry.v1.RegistryService/GetService",
			"/plantx.registry.v1.RegistryService/ListServices",
			"/plantx.registry.v1.RegistryService/RegisterMicroApp",
			"/plantx.registry.v1.RegistryService/ListMicroApps",
			"/plantx.registry.v1.RegistryService/UpdateMicroApp",
			"/plantx.registry.v1.RegistryService/DeleteMicroApp",
			"/plantx.registry.v1.RegistryService/CreateMenu",
			"/plantx.registry.v1.RegistryService/ListMenus",
			"/plantx.registry.v1.RegistryService/UpdateMenu",
			"/plantx.registry.v1.RegistryService/DeleteMenu",
			"/plantx.registry.v1.RegistryService/ReorderMenus",
			"/plantx.registry.v1.RegistryService/SyncRoutes",
			"/plantx.registry.v1.RegistryService/GetRoutePolicy",
			"/plantx.registry.v1.RegistryService/RegisterApplication",
			"/plantx.registry.v1.RegistryService/ListApplications",
			"/plantx.registry.v1.RegistryService/GetApplication",
			"/plantx.registry.v1.RegistryService/UpdateApplication",
			"/plantx.registry.v1.RegistryService/DeleteApplication",
			"/plantx.registry.v1.RegistryService/GetApplicationMenus",
			"/plantx.registry.v1.RegistryService/GetApplicationMicroApps",
		},
		Tracing: server.TracingOptions{
			Enabled:     cfg.GetBool("tracing_enabled"),
			ServiceName: "registry-service",
		},
	})

	migrationDir := cfg.GetString("migrations_dir")
	if migrationDir == "" {
		migrationDir = "./migrations"
	}
	if err := migrate.Up(context.Background(), sqldb, migrationDir); err != nil {
		logger.Error("failed to apply migrations", kitlog.F("error", err))
		os.Exit(1)
	}

	repository := repo.NewPostgresRepo(sqldb)

	var registryOpts []app.RegistryOption
	if temporalClient, err := temporal.NewClient(); err != nil {
		logger.Warn("failed to create temporal client; lifecycle workflows will not be triggered", kitlog.F("error", err))
	} else {
		registryOpts = append(registryOpts, app.WithTemporalClient(temporalClient))
		logger.Info("temporal client created")
	}

	registry := app.NewRegistry(repository, registryOpts...)
	handler := grpcsrv.NewHandler(registry)
	api.RegisterRegistryServiceServer(srv.GRPC(), handler)
	reflection.Register(srv.GRPC())

	if err := srv.RegisterGateway(context.Background(), api.RegisterRegistryServiceHandler); err != nil {
		logger.Error("failed to register gateway", kitlog.F("error", err))
		os.Exit(1)
	}
	logger.Info("registry gateway registered")

	logger.Info("registry service starting",
		kitlog.F("grpc_port", grpcPort),
		kitlog.F("http_port", httpPort),
	)
	if err := srv.Run(context.Background()); err != nil {
		logger.Error("server failed", kitlog.F("error", err))
		os.Exit(1)
	}
}
