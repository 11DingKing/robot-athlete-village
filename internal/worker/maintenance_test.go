package worker

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"log/slog"
	"testing"
	"time"
)

func workerStore(t *testing.T) repository.Store {
	t.Helper()
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, e := storage.Open(context.Background(), "file::memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	return repository.NewSQLite(db)
}
func TestMaintenanceProcess(t *testing.T) {
	s := workerStore(t)
	now := time.Now().UTC()
	j, e := s.QueueMaintenance(context.Background(), 1, now)
	if e != nil {
		t.Fatal(e)
	}
	w := NewMaintenance(s, time.Millisecond, slog.Default())
	w.process(context.Background(), now.Add(time.Second))
	var status string
	if e = s.DB().QueryRow("SELECT status FROM maintenance_jobs WHERE id=?", j.ID).Scan(&status); e != nil || status != "done" {
		t.Fatalf("%s %v", status, e)
	}
}
func TestMaintenanceProcessRestoreFailureLeavesEquipmentAndJobForRetry(t *testing.T) {
	s := workerStore(t)
	now := time.Now().UTC()
	j, e := s.QueueMaintenance(context.Background(), 1, now)
	if e != nil {
		t.Fatal(e)
	}
	// Simulate the controller rejecting restore: drop equipment out of
	// 'maintenance' so the guarded restore inside the atomic step cannot match.
	if _, e = s.DB().Exec("UPDATE equipment SET status='ready' WHERE id=?", j.EquipmentID); e != nil {
		t.Fatal(e)
	}
	w := NewMaintenance(s, time.Millisecond, slog.Default())
	w.process(context.Background(), now.Add(time.Second))
	var jobStatus, eqStatus string
	if e = s.DB().QueryRow("SELECT status FROM maintenance_jobs WHERE id=?", j.ID).Scan(&jobStatus); e != nil {
		t.Fatal(e)
	}
	if jobStatus != "retry" {
		t.Fatalf("job want retry got %s", jobStatus)
	}
	if e = s.DB().QueryRow("SELECT status FROM equipment WHERE id=?", j.EquipmentID).Scan(&eqStatus); e != nil {
		t.Fatal(e)
	}
	if eqStatus != "ready" {
		t.Fatalf("equipment want ready got %s", eqStatus)
	}
}
func TestMaintenanceStopsOnCancel(t *testing.T) {
	s := workerStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() { NewMaintenance(s, time.Millisecond, slog.Default()).Run(ctx); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
}
