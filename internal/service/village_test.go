package service

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/audit"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	appErr "github.com/11DingKing/robot-athlete-village/internal/errors"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"testing"
)

func villageService(t *testing.T) (*Village, repository.Store) {
	t.Helper()
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, e := storage.Open(context.Background(), "file::memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	store := repository.NewSQLite(db)
	return NewVillage(store, audit.New(store)), store
}
func TestAdmitRequiresAdmin(t *testing.T) {
	v, _ := villageService(t)
	coach := domain.User{ID: 2, Role: domain.RoleCoach}
	if _, e := v.Admit(context.Background(), coach, 1, 1, "x"); e != appErr.ErrForbidden {
		t.Fatal(e)
	}
}
func TestAdmitAndClose(t *testing.T) {
	v, _ := villageService(t)
	admin := domain.User{ID: 1, Role: domain.RoleAdmin}
	st, e := v.Admit(context.Background(), admin, 1, 1, "x")
	if e != nil {
		t.Fatal(e)
	}
	if st.Status != domain.StayActive {
		t.Fatal(st.Status)
	}
	closed, e := v.CloseStay(context.Background(), admin, st.ID)
	if e != nil || closed.Status != domain.StayClosed {
		t.Fatalf("%+v %v", closed, e)
	}
}
func TestReserveConfirm(t *testing.T) {
	v, _ := villageService(t)
	coach := domain.User{ID: 2, Role: domain.RoleCoach}
	b, e := v.ReserveTraining(context.Background(), coach, 1, 1, "b")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = v.ConfirmTraining(context.Background(), coach, b.ID); e != nil {
		t.Fatal(e)
	}
}
func TestConfirmAuditFailureLeavesHeld(t *testing.T) {
	v, store := villageService(t)
	coach := domain.User{ID: 2, Role: domain.RoleCoach}
	b, e := v.ReserveTraining(context.Background(), coach, 1, 1, "ba")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = store.DB().Exec("DROP TABLE audit_events"); e != nil {
		t.Fatal(e)
	}
	if _, e = v.ConfirmTraining(context.Background(), coach, b.ID); e == nil {
		t.Fatal("want error when audit is unavailable")
	}
	var status string
	if e = store.DB().QueryRow("SELECT status FROM bookings WHERE id=?", b.ID).Scan(&status); e != nil {
		t.Fatal(e)
	}
	if domain.BookingStatus(status) != domain.BookingHeld {
		t.Fatalf("want held after audit failure, got %s", status)
	}
}
func TestCheckinIdempotent(t *testing.T) {
	v, store := villageService(t)
	coach := domain.User{ID: 2, Role: domain.RoleCoach}
	for i := 0; i < 2; i++ {
		if e := v.Checkin(context.Background(), coach, 1, "opening", "check-1"); e != nil {
			t.Fatal(e)
		}
	}
	var n int
	if e := store.DB().QueryRow("SELECT COUNT(*) FROM checkins WHERE idempotency_key='check-1'").Scan(&n); e != nil || n != 1 {
		t.Fatalf("count %d err %v", n, e)
	}
}
func TestEquipmentAssignment(t *testing.T) {
	v, _ := villageService(t)
	coach := domain.User{ID: 2, Role: domain.RoleCoach}
	e, err := v.Assign(context.Background(), coach, 1, 1)
	if err != nil || e.Status != domain.EquipmentInUse {
		t.Fatalf("%+v %v", e, err)
	}
	if _, err = v.Assign(context.Background(), coach, 1, 2); err != appErr.ErrConflict {
		t.Fatalf("conflict %v", err)
	}
}
func TestMaintenanceAuthorization(t *testing.T) {
	v, _ := villageService(t)
	coach := domain.User{ID: 2, Role: domain.RoleCoach}
	if e := v.StartMaintenance(context.Background(), coach, 1); e != appErr.ErrForbidden {
		t.Fatal(e)
	}
	admin := domain.User{ID: 1, Role: domain.RoleAdmin}
	if e := v.StartMaintenance(context.Background(), admin, 1); e != nil {
		t.Fatal(e)
	}
}
