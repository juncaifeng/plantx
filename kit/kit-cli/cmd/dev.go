// Package cmd provides the kit CLI subcommands.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

// Dev manages the local development environment.
func Dev(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: kit dev up|down|logs")
		os.Exit(1)
	}

	var cmdArgs []string
	switch args[0] {
	case "up":
		cmdArgs = []string{"docker-compose", "-f", "deployments/docker-compose/docker-compose.yml", "up", "-d"}
	case "down":
		cmdArgs = []string{"docker-compose", "-f", "deployments/docker-compose/docker-compose.yml", "down"}
	case "logs":
		cmdArgs = []string{"docker-compose", "-f", "deployments/docker-compose/docker-compose.yml", "logs", "-f"}
	default:
		fmt.Fprintf(os.Stderr, "unknown dev command: %s\n", args[0])
		os.Exit(1)
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "dev command failed: %v\n", err)
		os.Exit(1)
	}
}
