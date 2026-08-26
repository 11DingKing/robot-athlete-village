package service

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"testing"
)

func TestAdmissionRollsBackWhenAuditFails(t *testing.T) {
	village, store := villageService(t)
	admin := domain.User{ID: 1, Role: domain.RoleAdmin}
	if _, err := store.DB().Exec(`
		CREATE TRIGGER reject_admission_audit
		BEFORE INSERT ON audit_events
		WHEN NEW.action = 'admit'
		BEGIN
			SELECT RAISE(FAIL, 'audit sink unavailable');
		END`); err != nil {
		t.Fatal(err)
	}

	_, err := village.Admit(context.Background(), admin, 1, 1, "audit-admit")
	if err == nil {
		t.Fatal("admission unexpectedly succeeded while mandatory audit was unavailable")
	}
	var stayCount, occupied int
	if err = store.DB().QueryRow("SELECT COUNT(*) FROM stays WHERE idempotency_key='audit-admit'").Scan(&stayCount); err != nil {
		t.Fatal(err)
	}
	if err = store.DB().QueryRow("SELECT occupied FROM rooms WHERE id=1").Scan(&occupied); err != nil {
		t.Fatal(err)
	}
	if stayCount != 0 || occupied != 0 {
		t.Fatalf("failed admission left persistent state: stays=%d occupied=%d", stayCount, occupied)
	}
	_ = domain.StayActive
}
