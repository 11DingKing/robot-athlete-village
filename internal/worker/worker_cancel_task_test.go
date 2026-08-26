package worker

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"log/slog"
	"testing"
	"time"
)

func TestStoppedWorkerLeavesQueuedMaintenanceUntouched(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, err := storage.Open(context.Background(), "file::memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := repository.NewSQLite(db)
	now := time.Now().UTC()
	job, err := store.QueueMaintenance(context.Background(), 1, now)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	NewMaintenance(store, time.Second, slog.Default()).process(ctx, now.Add(time.Second))
	var status string
	var attempts int
	if err := db.QueryRow("SELECT status, attempts FROM maintenance_jobs WHERE id=?", job.ID).Scan(&status, &attempts); err != nil {
		t.Fatal(err)
	}
	if status != "queued" || attempts != 0 {
		t.Fatalf("stopped worker changed job to status=%s attempts=%d", status, attempts)
	}
}
