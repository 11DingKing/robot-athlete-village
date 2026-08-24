package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
)

func TestFailedAdmissionKeepsRoomCapacity(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, err := storage.Open(context.Background(), "file::memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := repository.NewSQLite(db)

	_, err = store.CreateStay(context.Background(), 9999, 1, "invalid-delegation", time.Now().UTC())
	if err == nil {
		t.Fatal("admission with unknown delegation unexpectedly succeeded")
	}

	var occupied, version int
	if err := db.QueryRow("SELECT occupied, version FROM rooms WHERE id=1").Scan(&occupied, &version); err != nil {
		t.Fatal(err)
	}
	if occupied != 0 || version != 1 {
		t.Fatalf("failed admission changed room to occupied=%d version=%d, want 0/1", occupied, version)
	}

	stay, err := store.CreateStay(context.Background(), 1, 1, "valid-delegation", time.Now().UTC())
	if err != nil {
		t.Fatalf("valid admission after rollback failed: %v", err)
	}
	if stay.RoomID != 1 {
		t.Fatalf("valid admission room=%d, want 1", stay.RoomID)
	}
}
