package storage

import (
	"context"
	"database/sql"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"testing"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, err := Open(context.Background(), "file::memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
func TestMigrationsCreateTables(t *testing.T) {
	db := openTestDB(t)
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n < 14 {
		t.Fatalf("tables %d", n)
	}
}
func TestMigrationsAreIdempotent(t *testing.T) {
	db := openTestDB(t)
	if err := ApplyMigrations(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("versions %d", n)
	}
}
func TestSeedDataPresent(t *testing.T) {
	db := openTestDB(t)
	var users, rooms, venues int
	_ = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&users)
	_ = db.QueryRow("SELECT COUNT(*) FROM rooms").Scan(&rooms)
	_ = db.QueryRow("SELECT COUNT(*) FROM venues").Scan(&venues)
	if users < 2 || rooms < 3 || venues < 2 {
		t.Fatalf("seed %d %d %d", users, rooms, venues)
	}
}
func TestForeignKeysEnabled(t *testing.T) {
	db := openTestDB(t)
	if _, err := db.Exec("INSERT INTO athletes(delegation_id,display_name,category,status,created_at) VALUES(999,'x','x','ready','now')"); err == nil {
		t.Fatal("foreign key disabled")
	}
}
func TestRepositoryConstructs(t *testing.T) {
	db := openTestDB(t)
	if repository.NewSQLite(db).DB() != db {
		t.Fatal("db mismatch")
	}
}
