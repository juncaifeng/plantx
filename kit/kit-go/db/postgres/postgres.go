// Package postgres provides a PostgreSQL implementation of db.DB.
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/plantx/kit/kit-go/db"
)

// New opens a PostgreSQL connection pool and returns a db.DB wrapper.
func New(dsn string) (db.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is empty")
	}
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	sqldb.SetMaxOpenConns(25)
	sqldb.SetMaxIdleConns(5)
	return &wrapper{db: sqldb}, nil
}

type wrapper struct {
	db *sql.DB
}

func (w *wrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return w.db.QueryContext(ctx, query, args...)
}

func (w *wrapper) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return w.db.QueryRowContext(ctx, query, args...)
}

func (w *wrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return w.db.ExecContext(ctx, query, args...)
}

func (w *wrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return w.db.PrepareContext(ctx, query)
}

func (w *wrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return w.db.BeginTx(ctx, opts)
}

func (w *wrapper) PingContext(ctx context.Context) error {
	return w.db.PingContext(ctx)
}

func (w *wrapper) Close() error {
	return w.db.Close()
}
