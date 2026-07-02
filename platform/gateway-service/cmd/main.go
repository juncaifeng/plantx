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
	"github.com/plantx/platform/gateway-service/api"
	"github.com/plantx/platform/gateway-service/internal/app"
	"github.com/plantx/platform/gateway-service/internal/infra/registry"
	grpcsrv "github.com/plantx/platform/gateway-service/internal/interfaces/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	stdLogger := log.New(os.Stderr, "gateway-service: ", log.LstdFlags)

	logger, err := zaplog.New()
	if err != nil {
		stdLogger.Printf("failed to create logger: %v", err)
		os.Exit(1)
	}

	cfg := env.New("GATEWAY")
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
		GRPCPort:       grpcPort,
		HTTPPort:       httpPort,
		Authenticator:  authenticator,
		Authorizer:     authorizer,
		TenantResolver: tenant.NewResolver(),
		Logger:         logger,
		Config:         cfg,
		ServiceName:    "gateway-service",
		GatewayRegistrar: gateway.AutoRegister("gateway-service",
			gateway.WithApplication(gateway.Application{
				Key:      "platform",
				Name:     "Platform",
				LabelKey: "nav.admin",
			}),
			gateway.WithMicroApp(gateway.MicroApp{
				Name:              "gateway-admin-ui",
				Route:             "/admin/gateway",
				BundleURL:         "/apps/gateway-admin-ui/gateway-admin-ui.js",
				MenuLabelKey:      "nav.gateway",
				RequirePermission: "gateway:read",
			}),
		),
		AuthExcludedMethods: []string{
			"/plantx.gateway.v1.GatewayService/RegisterService",
			"/plantx.gateway.v1.GatewayService/RegisterMicroApp",
			"/plantx.gateway.v1.GatewayService/ListServices",
			"/plantx.gateway.v1.GatewayService/ListMicroApps",
		},
		Tracing: server.TracingOptions{
			Enabled:     cfg.GetBool("tracing_enabled"),
			ServiceName: "gateway-service",
		},
	})

	registryAddr := os.Getenv("REGISTRY_SERVICE_GRPC_ADDR")
	if registryAddr == "" {
		registryAddr = "registry-service:8080"
	}
	repository, err := registry.NewClient(registryAddr)
	if err != nil {
		logger.Error("failed to connect to registry-service", kitlog.F("error", err))
		os.Exit(1)
	}
	defer func() { _ = repository.Close() }()

	reg := app.NewRegistry(repository)
	handler := grpcsrv.NewHandler(reg)
	api.RegisterGatewayServiceServer(srv.GRPC(), handler)
	reflection.Register(srv.GRPC())

	if err := srv.RegisterGateway(context.Background(), api.RegisterGatewayServiceHandler); err != nil {
		logger.Error("failed to register gateway", kitlog.F("error", err))
		os.Exit(1)
	}
	logger.Info("gateway gateway registered")

	logger.Info("gateway service starting",
		kitlog.F("grpc_port", grpcPort),
		kitlog.F("http_port", httpPort),
	)
	if err := srv.Run(context.Background()); err != nil {
		logger.Error("server failed", kitlog.F("error", err))
		os.Exit(1)
	}
}
