package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestReopenPreservesPersistedAuditHistory(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	databasePath := filepath.Join(t.TempDir(), "village.db")
	first, err := Open(context.Background(), databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = first.Exec(`INSERT INTO audit_events(actor_user_id,entity_type,entity_id,action,result,request_id,created_at)
		VALUES(1,'equipment','1','inspect','success','restart-proof','2026-08-24T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := Open(context.Background(), databasePath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Close() })
	var count int
	if err = second.QueryRow("SELECT COUNT(*) FROM audit_events WHERE request_id='restart-proof'").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("persisted audit history lost after reopen: count=%d", count)
	}
}
