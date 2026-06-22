// Package migrate provides a minimal PostgreSQL migration runner for registry-service.
package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/plantx/kit/kit-go/db"
)

// Up applies all pending up migrations from migrationDir.
func Up(ctx context.Context, database db.DB, migrationDir string) error {
	if _, err := database.ExecContext(ctx, createMigrationsTable); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	files, err := listMigrationFiles(migrationDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		version := migrationVersion(file)
		applied, err := isApplied(ctx, database, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		data, err := os.ReadFile(filepath.Join(migrationDir, file))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		if _, err := database.ExecContext(ctx, string(data)); err != nil {
			return fmt.Errorf("apply migration %s: %w", file, err)
		}
		if _, err := database.ExecContext(ctx, insertMigration, version); err != nil {
			return fmt.Errorf("record migration %s: %w", file, err)
		}
	}
	return nil
}

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

const insertMigration = `INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING`

func isApplied(ctx context.Context, database db.DB, version string) (bool, error) {
	var exists bool
	row := database.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version)
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return exists, nil
}

func listMigrationFiles(migrationDir string) ([]string, error) {
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)
	return files, nil
}

func migrationVersion(file string) string {
	base := path.Base(file)
	base = strings.TrimSuffix(base, ".up.sql")
	return base
}

// DownFS is unused but kept for fs.FS compatibility in tests.
func DownFS() fs.FS {
	return os.DirFS(".")
}
