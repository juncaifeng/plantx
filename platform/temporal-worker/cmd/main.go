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
	if err := worker.Start(context.Background(), host); err != nil {
		log.Fatalf("worker failed: %v", err)
	}
}
