package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Migrate creates new database migration files.
func Migrate(args []string) {
	if len(args) < 2 || args[0] != "new" {
		fmt.Fprintln(os.Stderr, "usage: kit migrate new <name>")
		os.Exit(1)
	}

	name := args[1]
	ts := time.Now().Format("20060102150405")
	base := fmt.Sprintf("migrations/%s_%s", ts, name)

	if err := os.MkdirAll("migrations", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create migrations dir: %v\n", err)
		os.Exit(1)
	}

	upFile := base + ".up.sql"
	downFile := base + ".down.sql"

	if err := os.WriteFile(upFile, []byte("-- migration up\n"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create up migration: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(downFile, []byte("-- migration down\n"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create down migration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created migrations:\n  %s\n  %s\n", filepath.Clean(upFile), filepath.Clean(downFile))
}
