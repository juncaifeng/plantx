package main

import (
	"context"
	"log"
	"os"

	"github.com/plantx/kit/kit-go/config/env"
	"github.com/plantx/kit/kit-go/event"
	eventnats "github.com/plantx/kit/kit-go/event/nats"
	"github.com/plantx/kit/kit-go/gateway"
	kitlog "github.com/plantx/kit/kit-go/log"
	zaplog "github.com/plantx/kit/kit-go/log/zap"
	"github.com/plantx/kit/kit-go/server"
	"github.com/plantx/kit/kit-go/tenant"
	"github.com/plantx/platform/audit-service/api"
	"github.com/plantx/platform/audit-service/internal/app"
	auditevent "github.com/plantx/platform/audit-service/internal/event"
	"github.com/plantx/platform/audit-service/internal/infra/repo"
	grpcsrv "github.com/plantx/platform/audit-service/internal/interfaces/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	stdLogger := log.New(os.Stderr, "audit-service: ", log.LstdFlags)

	logger, err := zaplog.New()
	if err != nil {
		stdLogger.Printf("failed to create logger: %v", err)
		os.Exit(1)
	}

	cfg := env.New("AUDIT")
	grpcPort := cfg.GetInt("grpc_port")
	if grpcPort == 0 {
		grpcPort = 8080
	}
	httpPort := cfg.GetInt("http_port")
	if httpPort == 0 {
		httpPort = 8081
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

	repository := repo.NewInMemoryRepo()
	auditApp := app.NewAuditService(repository)
	handler := grpcsrv.NewHandler(auditApp)

	sub := auditevent.NewSubscriber(auditApp)
	if err := eventBus.Subscribe("audit.events", sub.Handle); err != nil {
		logger.Error("failed to subscribe to audit events", kitlog.F("error", err))
		os.Exit(1)
	}

	srv := server.New(server.Options{
		ServiceName:    "audit-service",
		GRPCPort:       grpcPort,
		HTTPPort:       httpPort,
		TenantResolver: tenant.NewResolver(),
		Logger:         logger,
		Config:         cfg,
		EventBus:       eventBus,
		GatewayRegistrar: gateway.AutoRegister("audit-service",
			gateway.WithApplication(gateway.Application{
				Key:      "platform",
				Name:     "Platform",
				LabelKey: "nav.admin",
			}),
			gateway.WithMicroApp(gateway.MicroApp{
				Name:              "audit-admin-ui",
				Route:             "/admin/audit",
				BundleURL:         "/apps/audit-admin-ui/audit-admin-ui.js",
				MenuLabelKey:      "nav.audit",
				RequirePermission: "audit:read",
			}),
		),
		Tracing: server.TracingOptions{
			Enabled:     cfg.GetBool("tracing_enabled"),
			ServiceName: "audit-service",
		},
	})

	api.RegisterAuditServiceServer(srv.GRPC(), handler)
	reflection.Register(srv.GRPC())

	if err := srv.RegisterGateway(context.Background(), api.RegisterAuditServiceHandler); err != nil {
		logger.Error("failed to register gateway", kitlog.F("error", err))
		os.Exit(1)
	}
	logger.Info("audit gateway registered")

	logger.Info("audit service starting",
		kitlog.F("grpc_port", grpcPort),
		kitlog.F("http_port", httpPort),
	)
	if err := srv.Run(context.Background()); err != nil {
		logger.Error("server failed", kitlog.F("error", err))
		os.Exit(1)
	}
}
