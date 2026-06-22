package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Build builds container images for the current service.
func Build(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	tag := fs.String("tag", "latest", "image tag")
	_ = fs.Parse(args)

	serviceDir := "."
	imageName := fmt.Sprintf("plantx/%s:%s", filepath.Base(serviceDir), *tag)

	cmd := exec.Command("docker", "build", "-t", imageName, serviceDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("built image %s\n", imageName)
}
