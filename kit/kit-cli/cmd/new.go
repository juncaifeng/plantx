package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/plantx/kit/kit-cli/internal/scaffold"
)

// New scaffolds a new service.
func New(args []string) {
	if len(args) < 2 || args[0] != "service" {
		fmt.Fprintln(os.Stderr, "usage: kit new service <name> [--ui] [--gateway]")
		os.Exit(1)
	}

	name := args[1]
	fs := flag.NewFlagSet("new service", flag.ExitOnError)
	withUI := fs.Bool("ui", false, "include web sub-application")
	withGateway := fs.Bool("gateway", false, "include grpc-gateway")
	_ = fs.Parse(args[2:])

	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find project root: %v\n", err)
		os.Exit(1)
	}
	if err := os.Chdir(root); err != nil {
		fmt.Fprintf(os.Stderr, "failed to change to project root: %v\n", err)
		os.Exit(1)
	}

	if err := scaffold.Service(name, scaffold.ServiceOptions{
		WithUI:      *withUI,
		WithGateway: *withGateway,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to scaffold service: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created service %s at %s\n", name, filepath.Join("services", name))
}
