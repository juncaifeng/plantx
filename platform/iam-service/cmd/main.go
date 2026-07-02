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
	"github.com/plantx/platform/iam-service/api"
	"github.com/plantx/platform/iam-service/internal/app"
	"github.com/plantx/platform/iam-service/internal/domain"
	"github.com/plantx/platform/iam-service/internal/infra/repo"
	grpcsrv "github.com/plantx/platform/iam-service/internal/interfaces/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	stdLogger := log.New(os.Stderr, "iam-service: ", log.LstdFlags)

	logger, err := zaplog.New()
	if err != nil {
		stdLogger.Printf("failed to create logger: %v", err)
		os.Exit(1)
	}

	cfg := env.New("IAM")
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

	var sqldb db.DB
	var repository domain.Repository
	if dsn := cfg.GetString("database_dsn"); dsn != "" {
		sqldb, err = postgres.New(dsn)
		if err != nil {
			logger.Error("failed to connect to postgres; falling back to in-memory", kitlog.F("error", err))
		} else {
			repository = repo.NewPostgresRepo(sqldb)
		}
	}
	if repository == nil {
		repository = repo.NewInMemoryRepo()
		logger.Info("using in-memory repository")
	}

	readiness := func(ctx context.Context) error {
		if sqldb != nil {
			return sqldb.PingContext(ctx)
		}
		return nil
	}

	srv := server.New(server.Options{
		ServiceName:    "iam-service",
		GRPCPort:       grpcPort,
		HTTPPort:       httpPort,
		Authenticator:  authenticator,
		Authorizer:     authorizer,
		TenantResolver: tenant.NewResolver(),
		Logger:         logger,
		Config:         cfg,
		DB:             sqldb,
		Readiness:      readiness,
		AuthExcludedMethods: []string{
			// Allow services to self-register their permission catalog and
			// allow the public/portal to read the permission catalog without
			// requiring an authenticated session.
			"/plantx.iam.v1.IAMService/CreatePermission",
			"/plantx.iam.v1.IAMService/ListPermissions",
		},
		GatewayRegistrar: gateway.AutoRegister("iam-service",
			gateway.WithApplication(gateway.Application{
				Key:      "platform",
				Name:     "Platform",
				LabelKey: "nav.admin",
			}),
			gateway.WithMicroApp(gateway.MicroApp{Name: "iam-admin-ui", Route: "/admin/iam", BundleURL: "/apps/iam-admin-ui/iam-admin-ui.js", MenuLabelKey: "nav.iam", RequirePermission: "iam:read"}),
		),
		Tracing: server.TracingOptions{
			Enabled:     cfg.GetBool("tracing_enabled"),
			ServiceName: "iam-service",
		},
	})

	iamApp := app.NewIAMService(repository)
	handler := grpcsrv.NewHandler(iamApp)
	api.RegisterIAMServiceServer(srv.GRPC(), handler)
	reflection.Register(srv.GRPC())

	if err := srv.RegisterGateway(context.Background(), api.RegisterIAMServiceHandler); err != nil {
		logger.Error("failed to register gateway", kitlog.F("error", err))
		os.Exit(1)
	}
	logger.Info("iam gateway registered")

	logger.Info("iam service starting",
		kitlog.F("grpc_port", grpcPort),
		kitlog.F("http_port", httpPort),
	)
	if err := srv.Run(context.Background()); err != nil {
		logger.Error("server failed", kitlog.F("error", err))
		os.Exit(1)
	}
}
