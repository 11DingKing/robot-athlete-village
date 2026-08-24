package repository

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	appErr "github.com/11DingKing/robot-athlete-village/internal/errors"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"testing"
	"time"
)

func newRepo(t *testing.T) *SQLite {
	t.Helper()
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, e := storage.Open(context.Background(), "file::memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	return NewSQLite(db)
}
func TestFindUser(t *testing.T) {
	r := newRepo(t)
	u, e := r.FindUser(context.Background(), "mayor@example.com")
	if e != nil || u.Role != domain.RoleAdmin || !u.Active {
		t.Fatalf("%+v %v", u, e)
	}
	if _, e = r.FindUser(context.Background(), "missing@example.com"); e != appErr.ErrNotFound {
		t.Fatalf("want not found got %v", e)
	}
}
func TestSessionLifecycle(t *testing.T) {
	r := newRepo(t)
	u, _ := r.FindUser(context.Background(), "coach@example.com")
	if e := r.CreateSession(context.Background(), u, "tok", time.Now().Add(time.Hour)); e != nil {
		t.Fatal(e)
	}
	got, _, e := r.FindSession(context.Background(), "tok")
	if e != nil || got.ID != u.ID {
		t.Fatalf("%+v %v", got, e)
	}
	if e = r.RevokeSession(context.Background(), "tok"); e != nil {
		t.Fatal(e)
	}
	if _, _, e = r.FindSession(context.Background(), "tok"); e != appErr.ErrUnauthorized {
		t.Fatalf("revoked %v", e)
	}
}
func TestStayCapacityAndClose(t *testing.T) {
	r := newRepo(t)
	now := time.Now().UTC()
	a, e := r.CreateStay(context.Background(), 1, 2, "k1", now)
	if e != nil || a.Status != domain.StayActive {
		t.Fatalf("%+v %v", a, e)
	}
	if _, e = r.CreateStay(context.Background(), 1, 2, "k2", now); e != nil {
		t.Fatal(e)
	}
	if _, e = r.CreateStay(context.Background(), 1, 2, "k3", now); e != appErr.ErrCapacity {
		t.Fatalf("capacity %v", e)
	}
	closed, e := r.TransitionStay(context.Background(), a.ID, domain.StayActive, domain.StayClosed, now)
	if e != nil || closed.Status != domain.StayClosed {
		t.Fatalf("%+v %v", closed, e)
	}
}
func TestStayIdempotency(t *testing.T) {
	r := newRepo(t)
	now := time.Now().UTC()
	a, e := r.CreateStay(context.Background(), 1, 1, "same", now)
	if e != nil {
		t.Fatal(e)
	}
	b, e := r.CreateStay(context.Background(), 1, 1, "same", now)
	if e != nil || a.ID != b.ID {
		t.Fatalf("%+v %+v %v", a, b, e)
	}
}
func TestBookingLifecycle(t *testing.T) {
	r := newRepo(t)
	b, e := r.CreateBooking(context.Background(), 1, 1, 2, "bk")
	if e != nil || b.Status != domain.BookingHeld {
		t.Fatalf("%+v %v", b, e)
	}
	b, e = r.TransitionBooking(context.Background(), b.ID, domain.BookingHeld, domain.BookingConfirmed)
	if e != nil || b.Status != domain.BookingConfirmed {
		t.Fatalf("%+v %v", b, e)
	}
	if _, e = r.TransitionBooking(context.Background(), b.ID, domain.BookingHeld, domain.BookingCancelled); e != appErr.ErrConflict {
		t.Fatalf("conflict %v", e)
	}
}
func TestRevertStayReleasesRoom(t *testing.T) {
	r := newRepo(t)
	now := time.Now().UTC()
	st, e := r.CreateStay(context.Background(), 1, 2, "rev", now)
	if e != nil {
		t.Fatal(e)
	}
	if e = r.RevertStay(context.Background(), st.ID, st.RoomID, st.Version); e != nil {
		t.Fatal(e)
	}
	var n, occ int
	_ = r.DB().QueryRow("SELECT COUNT(*) FROM stays WHERE id=?", st.ID).Scan(&n)
	_ = r.DB().QueryRow("SELECT occupied FROM rooms WHERE id=?", st.RoomID).Scan(&occ)
	if n != 0 || occ != 0 {
		t.Fatalf("stays %d occupied %d", n, occ)
	}
	if e = r.RevertStay(context.Background(), st.ID, st.RoomID, st.Version); e != appErr.ErrConflict {
		t.Fatalf("want conflict on stale revert got %v", e)
	}
}
