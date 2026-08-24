package storage

import (
	"context"
	"path/filepath"
	"testing"
)

// TestOpenPreservesFileDatabaseAcrossReopens reproduces the restart bug:
// a process writes audit history to a file DB, restarts, and must still
// see the previous rows instead of an empty in-memory database.
func TestOpenPreservesFileDatabaseAcrossReopens(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	dir := t.TempDir()
	dsn := "file:" + filepath.Join(dir, "restart-audit.db")

	first, err := Open(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open first: %v", err)
	}
	if _, err = first.Exec(`INSERT INTO audit_events(entity_type,entity_id,action,result,request_id,created_at)
		VALUES('stay','1','open','success','r1','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err = first.Close(); err != nil {
		t.Fatalf("close first: %v", err)
	}

	// Simulate restart: the file now exists on disk. The old code redirected
	// to file::memory: here, hiding the previously written audit history.
	second, err := Open(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open second: %v", err)
	}
	t.Cleanup(func() { second.Close() })

	var n int
	if err = second.QueryRow("SELECT COUNT(*) FROM audit_events").Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 1 {
		t.Fatalf("audit history disappeared after restart: got %d rows, want 1", n)
	}
}
