package service

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"testing"
	"time"
)

func operationsService(t *testing.T) *Operations {
	t.Helper()
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, e := storage.Open(context.Background(), "file::memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	return NewOperations(repository.NewSQLite(db), DefaultPolicy())
}
func TestValidateAthleteOperation(t *testing.T) {
	o := operationsService(t)
	if e := o.ValidateAthlete(context.Background(), 1); e != nil {
		t.Fatal(e)
	}
	if e := o.ValidateAthlete(context.Background(), 999); e == nil {
		t.Fatal("missing accepted")
	}
}
func TestCategoryOperation(t *testing.T) {
	o := operationsService(t)
	if !o.IsCategoryAllowed("mobility") || o.IsCategoryAllowed("unknown") {
		t.Fatal("category")
	}
}
func TestReserveLimit(t *testing.T) {
	o := operationsService(t)
	u := domain.User{ID: 2, Role: domain.RoleCoach}
	if _, e := o.ReserveIfAllowed(context.Background(), u, domain.BookingRequest{AthleteID: 1, SlotID: 1, CoachID: 2, IdempotencyKey: "op"}, 3); e == nil {
		t.Fatal("limit ignored")
	}
	if _, e := o.ReserveIfAllowed(context.Background(), u, domain.BookingRequest{AthleteID: 1, SlotID: 1, CoachID: 2, IdempotencyKey: "op2"}, 0); e != nil {
		t.Fatal(e)
	}
}
func TestCloseWindow(t *testing.T) {
	o := operationsService(t)
	u := domain.User{ID: 1, Role: domain.RoleAdmin}
	base := time.Now().UTC()
	if _, e := o.CloseIfEligible(context.Background(), u, 1, base, base.Add(time.Hour)); e == nil {
		t.Fatal("early close")
	}
}
