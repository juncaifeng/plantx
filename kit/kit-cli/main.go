package main

import (
	"fmt"
	"os"

	"github.com/plantx/kit/kit-cli/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "new":
		cmd.New(args)
	case "generate", "gen":
		cmd.Generate(args)
	case "migrate":
		cmd.Migrate(args)
	case "dev":
		cmd.Dev(args)
	case "test":
		cmd.Test(args)
	case "build":
		cmd.Build(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`kit - PlantX Kit CLI

Usage:
  kit <command> [args]

Commands:
  new service <name> [--ui] [--gateway]   Scaffold a new service
  generate                                Generate code from proto and sqlc
  migrate new <name>                      Create a new database migration
  dev up|down|logs                        Manage local development environment
  test                                    Run tests
  build --tag <tag>                       Build container images
  help                                    Show this help message
`)
}
