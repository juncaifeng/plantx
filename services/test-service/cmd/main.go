package main

import (
	"context"
	"log"
	"os"

	"github.com/plantx/kit/kit-go/auth/maxkey"
	"github.com/plantx/kit/kit-go/config/env"
	"github.com/plantx/kit/kit-go/gateway"
	kitlog "github.com/plantx/kit/kit-go/log"
	zaplog "github.com/plantx/kit/kit-go/log/zap"
	"github.com/plantx/kit/kit-go/server"
	"github.com/plantx/services/test-service/api"
	"github.com/plantx/services/test-service/internal/app"
	grpcsrv "github.com/plantx/services/test-service/internal/interfaces/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	stdLogger := log.New(os.Stderr, "test-service: ", log.LstdFlags)

	logger, err := zaplog.New()
	if err != nil {
		stdLogger.Printf("failed to create logger: %v", err)
		os.Exit(1)
	}

	cfg := env.New("TEST")
	grpcPort := cfg.GetInt("grpc_port")
	if grpcPort == 0 {
		grpcPort = 8080
	}
	httpPort := cfg.GetInt("http_port")
	if httpPort == 0 {
		httpPort = 8081
	}

	srv := server.New(server.Options{
		ServiceName:   "test-service",
		GRPCPort:      grpcPort,
		HTTPPort:      httpPort,
		Logger:        logger,
		Config:        cfg,
		Authenticator: maxkey.New(maxkey.EnvOptions("MAXKEY")),
		GatewayRegistrar: gateway.AutoRegister("test-service",
			gateway.WithApplication(gateway.Application{
				Key:      "test",
				Name:     "Test",
				LabelKey: "nav.test",
			}),
			gateway.WithMicroApp(gateway.MicroApp{
				Name:              "test-ui",
				Route:             "/test",
				BundleURL:         "/apps/test-ui/test-ui.js",
				MenuLabelKey:      "nav.test",
				RequirePermission: "test:read",
			}),
		),
	})

	testApp := app.New()
	handler := grpcsrv.NewHandler(testApp)
	api.RegisterTestServiceServer(srv.GRPC(), handler)
	reflection.Register(srv.GRPC())

	if err := srv.RegisterGateway(context.Background(), api.RegisterTestServiceHandler); err != nil {
		logger.Error("failed to register gateway", kitlog.F("error", err))
		os.Exit(1)
	}
	logger.Info("test gateway registered")

	logger.Info("test service starting",
		kitlog.F("grpc_port", grpcPort),
		kitlog.F("http_port", httpPort),
	)
	if err := srv.Run(context.Background()); err != nil {
		logger.Error("server failed", kitlog.F("error", err))
		os.Exit(1)
	}
}
