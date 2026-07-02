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
	"github.com/plantx/kit/kit-go/event"
	eventnats "github.com/plantx/kit/kit-go/event/nats"
	"github.com/plantx/kit/kit-go/gateway"
	kitlog "github.com/plantx/kit/kit-go/log"
	zaplog "github.com/plantx/kit/kit-go/log/zap"
	"github.com/plantx/kit/kit-go/server"
	"github.com/plantx/kit/kit-go/tenant"
	"github.com/plantx/services/order/api"
	"github.com/plantx/services/order/internal/app"
	"github.com/plantx/services/order/internal/domain"
	"github.com/plantx/services/order/internal/infra/repo"
	grpcsrv "github.com/plantx/services/order/internal/interfaces/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	stdLogger := log.New(os.Stderr, "order-service: ", log.LstdFlags)

	logger, err := zaplog.New()
	if err != nil {
		stdLogger.Printf("failed to create logger: %v", err)
		os.Exit(1)
	}

	cfg := env.New("ORDER")
	grpcPort := cfg.GetInt("grpc_port")
	if grpcPort == 0 {
		grpcPort = 8080
	}
	httpPort := cfg.GetInt("http_port")
	if httpPort == 0 {
		httpPort = 8081
	}

	var sqldb db.DB
	var repository domain.Repository
	if dsn := cfg.GetString("database_dsn"); dsn != "" {
		sqldb, err = postgres.New(dsn)
		if err != nil {
			logger.Error("failed to connect to postgres; falling back to in-memory", kitlog.F("error", err))
		} else {
			repository = repo.NewTenantRepo(repo.NewOrderRepo(sqldb))
		}
	}
	if repository == nil {
		repository = repo.NewTenantRepo(repo.NewInMemoryRepo())
		logger.Info("using in-memory repository")
	}

	var eventBus event.Bus = eventnats.NewInMemory()
	if natsURL := cfg.GetString("nats_url"); natsURL != "" {
		b, err := eventnats.New(eventnats.Options{URL: natsURL})
		if err != nil {
			logger.Error("failed to connect to nats; using in-memory bus", kitlog.F("error", err))
		} else {
			eventBus = b
		}
	}
	defer eventBus.Close()

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

	readiness := func(ctx context.Context) error {
		if sqldb != nil {
			return sqldb.PingContext(ctx)
		}
		return nil
	}

	srv := server.New(server.Options{
		ServiceName:    "order-service",
		GRPCPort:       grpcPort,
		HTTPPort:       httpPort,
		Authenticator:  authenticator,
		Authorizer:     authorizer,
		TenantResolver: tenant.NewResolver(),
		Logger:         logger,
		Config:         cfg,
		DB:             sqldb,
		EventBus:       eventBus,
		Tracing: server.TracingOptions{
			Enabled:     cfg.GetBool("tracing_enabled"),
			ServiceName: "order-service",
		},
		Readiness: readiness,
		GatewayRegistrar: gateway.AutoRegister("order-service",
			gateway.WithApplication(gateway.Application{
				Key:      "order",
				Name:     "Order",
				LabelKey: "nav.orders",
			}),
			gateway.WithMicroApp(gateway.MicroApp{
				Name:              "order-ui",
				Route:             "/order",
				BundleURL:         "/apps/order-ui/order-ui.js",
				MenuLabelKey:      "nav.orders",
				RequirePermission: "order:read",
			}),
		),
	})

	orderApp := app.NewOrderService(repository)
	handler := grpcsrv.NewHandler(orderApp)
	api.RegisterOrderServiceServer(srv.GRPC(), handler)
	reflection.Register(srv.GRPC())

	if err := srv.RegisterGateway(context.Background(), api.RegisterOrderServiceHandler); err != nil {
		logger.Error("failed to register gateway", kitlog.F("error", err))
		os.Exit(1)
	}
	logger.Info("order gateway registered")

	logger.Info("order service starting",
		kitlog.F("grpc_port", grpcPort),
		kitlog.F("http_port", httpPort),
	)
	if err := srv.Run(context.Background()); err != nil {
		logger.Error("server failed", kitlog.F("error", err))
		os.Exit(1)
	}
}
