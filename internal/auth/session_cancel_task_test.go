package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/11DingKing/robot-athlete-village/internal/auth"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
)

func TestCancelledLoginDoesNotCreateSession(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, err := storage.Open(context.Background(), "file::memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	service := auth.New(repository.NewSQLite(db), time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	token, _, err := service.Login(ctx, "coach@example.com", "coach-secret")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled login error = %v, want context.Canceled", err)
	}
	if token != "" {
		t.Fatalf("cancelled login returned token %q", token)
	}

	var sessions int
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessions); err != nil {
		t.Fatal(err)
	}
	if sessions != 0 {
		t.Fatalf("cancelled login persisted %d sessions, want 0", sessions)
	}
}
