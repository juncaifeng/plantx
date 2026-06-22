package main

import (
	"context"
	"fmt"

	"github.com/plantx/kit/kit-go/server"
)

func main() {
	srv := server.New(server.Options{GRPCPort: 8080})
	fmt.Printf("starting %s service on :8080\n", "demotmp")
	if err := srv.Run(context.Background()); err != nil {
		panic(err)
	}
}
