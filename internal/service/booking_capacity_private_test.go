package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type synchronizedCapacityStore struct {
	repository.Store
	arrived chan struct{}
	release <-chan struct{}
	writeMu sync.Mutex
}

func (s *synchronizedCapacityStore) BookingCapacity(ctx context.Context, slotID int64) (int, int, error) {
	active, capacity, err := s.Store.BookingCapacity(ctx, slotID)
	if err != nil {
		return 0, 0, err
	}
	s.arrived <- struct{}{}
	select {
	case <-s.release:
		return active, capacity, nil
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	}
}

func (s *synchronizedCapacityStore) CreateBooking(ctx context.Context, athleteID, slotID, coachID int64, key string) (domain.Booking, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.Store.CreateBooking(ctx, athleteID, slotID, coachID, key)
}

func TestConcurrentTrainingReservationsRespectCapacity(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, err := storage.Open(context.Background(), filepath.Join(t.TempDir(), "village.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err = db.Exec("UPDATE venues SET capacity=1 WHERE id=1"); err != nil {
		t.Fatal(err)
	}

	release := make(chan struct{})
	store := &synchronizedCapacityStore{
		Store:   repository.NewSQLite(db),
		arrived: make(chan struct{}, 2),
		release: release,
	}
	village := NewVillage(store, nil)
	coach := domain.User{ID: 2, Role: domain.RoleCoach}
	results := make(chan error, 2)

	for i, athleteID := range []int64{1, 2} {
		go func(i int, athleteID int64) {
			_, reserveErr := village.ReserveTraining(context.Background(), coach, athleteID, 1, fmt.Sprintf("parallel-%d", i))
			results <- reserveErr
		}(i, athleteID)
	}
	for i := 0; i < 2; i++ {
		select {
		case <-store.arrived:
		case <-time.After(2 * time.Second):
			t.Fatal("concurrent reservations did not reach the shared capacity snapshot")
		}
	}
	close(release)

	succeeded := 0
	for i := 0; i < 2; i++ {
		if reserveErr := <-results; reserveErr == nil {
			succeeded++
		}
	}
	var active int
	if err = db.QueryRow("SELECT COUNT(*) FROM bookings WHERE slot_id=1 AND status IN ('held','confirmed')").Scan(&active); err != nil {
		t.Fatal(err)
	}
	if succeeded > 1 || active > 1 {
		t.Fatalf("capacity violated: successful reservations=%d active bookings=%d capacity=1", succeeded, active)
	}
}
