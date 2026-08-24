package storage

import (
	"context"
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if err = ApplyMigrations(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)"); err != nil {
		return err
	}
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "migrations"
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		for _, candidate := range []string{"../../migrations", "../../../migrations"} {
			if fallback, e := os.ReadDir(candidate); e == nil {
				dir, entries, err = candidate, fallback, nil
				break
			}
		}
	}
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		var v int
		if _, e := fmt.Sscanf(entry.Name(), "%d_", &v); e != nil {
			continue
		}
		var exists int
		if e := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version=?", v).Scan(&exists); e != nil {
			return e
		}
		if exists > 0 {
			continue
		}
		body, e := os.ReadFile(filepath.Join(dir, entry.Name()))
		if e != nil {
			return e
		}
		tx, e := db.BeginTx(ctx, nil)
		if e != nil {
			return e
		}
		for _, stmt := range strings.Split(string(body), ";") {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if _, e = tx.ExecContext(ctx, stmt); e != nil {
				tx.Rollback()
				return fmt.Errorf("migration %s: %w", entry.Name(), e)
			}
		}
		if _, e = tx.ExecContext(ctx, "INSERT INTO schema_migrations(version,applied_at) VALUES (?,datetime('now'))", v); e != nil {
			tx.Rollback()
			return e
		}
		if e = tx.Commit(); e != nil {
			return e
		}
	}
	return nil
}
