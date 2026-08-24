package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/robot-athlete-village/internal/audit"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"path/filepath"
	"testing"
	"time"
)

func batchService(t *testing.T) *Batch {
	t.Helper()
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, e := storage.Open(context.Background(), "file:"+filepath.Join(t.TempDir(), "batch.db"))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	s := repository.NewSQLite(db)
	return NewBatch(NewVillage(s, audit.New(s)))
}
func TestAdmitManyResultsKeepOrder(t *testing.T) {
	b := batchService(t)
	u := domain.User{ID: 1, Role: domain.RoleAdmin}
	reqs := []domain.AdmissionRequest{{DelegationID: 1, RoomID: 1, IdempotencyKey: "a"}, {DelegationID: 1, RoomID: 2, IdempotencyKey: "b"}, {DelegationID: 0, RoomID: 1, IdempotencyKey: "bad"}}
	out := b.AdmitMany(context.Background(), u, reqs)
	if len(out) != 3 || (!out[0].Success && out[0].Message == "") || (!out[1].Success && out[1].Message == "") || out[2].Success {
		t.Fatalf("%+v", out)
	}
}
func TestBookAndConfirm(t *testing.T) {
	b := batchService(t)
	u := domain.User{ID: 2, Role: domain.RoleCoach}
	book, e := b.BookAndConfirm(context.Background(), u, domain.BookingRequest{AthleteID: 1, SlotID: 1, CoachID: 2, IdempotencyKey: "batch-book"})
	if e != nil || book.Status != domain.BookingConfirmed {
		t.Fatalf("%+v %v", book, e)
	}
}
func TestBookAndConfirmReleasesHeldSlotOnConfirmFailure(t *testing.T) {
	b := batchService(t)
	coach := domain.User{ID: 2, Role: domain.RoleCoach}
	// Slot 3 belongs to venue 2 (capacity 4). Fill it to the limit.
	var last int64
	for i := 0; i < 4; i++ {
		bk, e := b.village.ReserveTraining(context.Background(), coach, 1, 3, fmt.Sprintf("fill-%d", i))
		if e != nil {
			t.Fatal(e)
		}
		last = bk.ID
	}
	// Reserving again must hit capacity while all four are held.
	if _, e := b.village.ReserveTraining(context.Background(), coach, 1, 3, "overflow"); e == nil {
		t.Fatal("expected capacity exceeded")
	}
	// A failed confirmation leaves the booking held; compensating it must
	// cancel the held booking and free the slot for the next robot.
	if e := b.village.CompensateFailedConfirmation(context.Background(), coach, last); e != nil {
		t.Fatalf("compensate: %v", e)
	}
	if _, e := b.village.ReserveTraining(context.Background(), coach, 1, 3, "reclaim"); e != nil {
		t.Fatalf("reserve after compensate: %v", e)
	}
}
func TestCloseExpired(t *testing.T) {
	b := batchService(t)
	u := domain.User{ID: 1, Role: domain.RoleAdmin}
	for i := 0; i < 2; i++ {
		_, e := b.village.Admit(context.Background(), u, 1, int64(i+1), string(rune('a'+i)))
		if e != nil {
			t.Fatal(e)
		}
	}
	n := b.CloseExpired(context.Background(), u, []int64{1, 2, 99}, time.Now().UTC())
	if n != 2 {
		t.Fatalf("closed %d", n)
	}
}
