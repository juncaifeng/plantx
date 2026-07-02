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
	"github.com/plantx/kit/kit-go/gateway"
	kitlog "github.com/plantx/kit/kit-go/log"
	zaplog "github.com/plantx/kit/kit-go/log/zap"
	"github.com/plantx/kit/kit-go/server"
	"github.com/plantx/kit/kit-go/tenant"
	"github.com/plantx/platform/tenant-service/api"
	"github.com/plantx/platform/tenant-service/internal/app"
	"github.com/plantx/platform/tenant-service/internal/infra/repo"
	grpcsrv "github.com/plantx/platform/tenant-service/internal/interfaces/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	stdLogger := log.New(os.Stderr, "tenant-service: ", log.LstdFlags)

	logger, err := zaplog.New()
	if err != nil {
		stdLogger.Printf("failed to create logger: %v", err)
		os.Exit(1)
	}

	cfg := env.New("TENANT")
	grpcPort := cfg.GetInt("grpc_port")
	if grpcPort == 0 {
		grpcPort = 8080
	}
	httpPort := cfg.GetInt("http_port")
	if httpPort == 0 {
		httpPort = 8081
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
	iamCfg := env.New("IAM")
	authorizer = authzopa.New(authzopa.Options{
		URL:          opaCfg.GetString("url"),
		DecisionPath: opaCfg.GetString("decision_path"),
		IAMURL:       iamCfg.GetString("url"),
	})

	srv := server.New(server.Options{
		ServiceName:    "tenant-service",
		GRPCPort:       grpcPort,
		HTTPPort:       httpPort,
		Authenticator:  authenticator,
		Authorizer:     authorizer,
		TenantResolver: tenant.NewResolver(),
		Logger:         logger,
		Config:         cfg,
		GatewayRegistrar: gateway.AutoRegister("tenant-service",
			gateway.WithApplication(gateway.Application{
				Key:      "platform",
				Name:     "Platform",
				LabelKey: "nav.admin",
			}),
			gateway.WithMicroApp(gateway.MicroApp{
				Name:              "tenant-admin-ui",
				Route:             "/admin/tenants",
				BundleURL:         "/apps/tenant-admin-ui/tenant-admin-ui.js",
				MenuLabelKey:      "nav.tenants",
				RequirePermission: "tenant:read",
			}),
		),
		Tracing: server.TracingOptions{
			Enabled:     cfg.GetBool("tracing_enabled"),
			ServiceName: "tenant-service",
		},
	})

	repository := repo.NewInMemoryRepo()
	tenantApp := app.NewTenantService(repository)
	handler := grpcsrv.NewHandler(tenantApp)
	api.RegisterTenantServiceServer(srv.GRPC(), handler)
	reflection.Register(srv.GRPC())

	if err := srv.RegisterGateway(context.Background(), api.RegisterTenantServiceHandler); err != nil {
		logger.Error("failed to register gateway", kitlog.F("error", err))
		os.Exit(1)
	}
	logger.Info("tenant gateway registered")

	logger.Info("tenant service starting",
		kitlog.F("grpc_port", grpcPort),
		kitlog.F("http_port", httpPort),
	)
	if err := srv.Run(context.Background()); err != nil {
		logger.Error("server failed", kitlog.F("error", err))
		os.Exit(1)
	}
}
