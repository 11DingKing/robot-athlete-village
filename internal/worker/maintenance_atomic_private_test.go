package worker

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestMaintenanceCompletionRollsBackWhenEquipmentRestoreFails(t *testing.T) {
	store := workerStore(t)
	now := time.Now().UTC()
	job, err := store.QueueMaintenance(context.Background(), 1, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.DB().Exec(`
		CREATE TRIGGER reject_equipment_restore
		BEFORE UPDATE OF status ON equipment
		WHEN NEW.status = 'ready'
		BEGIN
			SELECT RAISE(FAIL, 'equipment controller unavailable');
		END`); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	NewMaintenance(store, time.Second, logger).process(context.Background(), now.Add(time.Second))

	var jobStatus, equipmentStatus string
	if err = store.DB().QueryRow("SELECT status FROM maintenance_jobs WHERE id=?", job.ID).Scan(&jobStatus); err != nil {
		t.Fatal(err)
	}
	if err = store.DB().QueryRow("SELECT status FROM equipment WHERE id=1").Scan(&equipmentStatus); err != nil {
		t.Fatal(err)
	}
	if jobStatus == "done" || equipmentStatus != "maintenance" {
		t.Fatalf("partial completion persisted: job=%s equipment=%s", jobStatus, equipmentStatus)
	}
}
