package audit

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"github.com/11DingKing/robot-athlete-village/internal/telemetry"
	"testing"
)

func TestLoggerAddsRequestID(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, e := storage.Open(context.Background(), "file::memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	s := repository.NewSQLite(db)
	l := New(s)
	ctx := telemetry.WithRequestID(context.Background(), "audit-request")
	if e = l.Record(ctx, 1, "stay", "7", "admit", "success"); e != nil {
		t.Fatal(e)
	}
	var got string
	if e = db.QueryRow("SELECT request_id FROM audit_events").Scan(&got); e != nil || got != "audit-request" {
		t.Fatalf("%q %v", got, e)
	}
}
func TestLoggerCanRecordFailure(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, e := storage.Open(context.Background(), "file::memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	s := repository.NewSQLite(db)
	if e := New(s).Record(context.Background(), 1, "stay", "8", "admit", "failed"); e != nil {
		t.Fatal(e)
	}
	var result string
	_ = db.QueryRow("SELECT result FROM audit_events").Scan(&result)
	if result != "failed" {
		t.Fatal(result)
	}
}
