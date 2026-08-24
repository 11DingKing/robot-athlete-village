package service

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"testing"
)

func TestConfirmTrainingRollsBackWhenAuditFails(t *testing.T) {
	village, store := villageService(t)
	coach := domain.User{ID: 2, Role: domain.RoleCoach}
	booking, err := village.ReserveTraining(context.Background(), coach, 1, 1, "audit-confirm")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.DB().Exec(`
		CREATE TRIGGER reject_confirmation_audit
		BEFORE INSERT ON audit_events
		WHEN NEW.action = 'confirm'
		BEGIN
			SELECT RAISE(FAIL, 'audit sink unavailable');
		END`); err != nil {
		t.Fatal(err)
	}

	if _, err = village.ConfirmTraining(context.Background(), coach, booking.ID); err == nil {
		t.Fatal("confirmation unexpectedly succeeded while mandatory audit was unavailable")
	}
	var status string
	if err = store.DB().QueryRow("SELECT status FROM bookings WHERE id=?", booking.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != string(domain.BookingHeld) {
		t.Fatalf("failed confirmation persisted booking status=%s, want held", status)
	}
}
