package service

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"testing"
)

func TestBookAndConfirmCancelsHeldReservationAfterConfirmFailure(t *testing.T) {
	batch := batchService(t)
	coach := domain.User{ID: 2, Role: domain.RoleCoach}
	if _, err := batch.village.store.DB().Exec(`
		CREATE TRIGGER reject_booking_confirmation
		BEFORE UPDATE OF status ON bookings
		WHEN NEW.status = 'confirmed'
		BEGIN
			SELECT RAISE(FAIL, 'training calendar unavailable');
		END`); err != nil {
		t.Fatal(err)
	}

	_, err := batch.BookAndConfirm(context.Background(), coach, domain.BookingRequest{AthleteID: 1, SlotID: 1, CoachID: 2, IdempotencyKey: "compensate-confirm"})
	if err == nil {
		t.Fatal("booking unexpectedly confirmed while confirmation store was unavailable")
	}
	var status string
	if err = batch.village.store.DB().QueryRow("SELECT status FROM bookings WHERE idempotency_key='compensate-confirm'").Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != string(domain.BookingCancelled) {
		t.Fatalf("failed batch left held booking=%s, want cancelled compensation", status)
	}
}
