package main

import (
	"context"
	"log"
	"os"

	"github.com/plantx/platform/temporal-worker/internal/worker"
)

func main() {
	host := os.Getenv("TEMPORAL_HOST")
	if host == "" {
		host = "localhost:7233"
	}
	registryAddr := os.Getenv("REGISTRY_SERVICE_GRPC_ADDR")
	if registryAddr == "" {
		registryAddr = "localhost:8080"
	}
	if err := worker.Start(context.Background(), host, registryAddr); err != nil {
		log.Fatalf("worker failed: %v", err)
	}
}
