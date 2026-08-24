package repository

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	appErr "github.com/11DingKing/robot-athlete-village/internal/errors"
	"testing"
	"time"
)

func TestEquipmentWorkflow(t *testing.T) {
	r := newRepo(t)
	e, err := r.AssignEquipment(context.Background(), 1, 1)
	if err != nil || e.ID != 1 {
		t.Fatalf("%+v %v", e, err)
	}
	if _, err = r.AssignEquipment(context.Background(), 1, 2); err != appErr.ErrConflict {
		t.Fatalf("want conflict %v", err)
	}
	j, err := r.QueueMaintenance(context.Background(), 1, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := r.ClaimMaintenance(context.Background(), time.Now().Add(time.Second))
	if err != nil || claimed.ID != j.ID {
		t.Fatalf("%+v %v", claimed, err)
	}
	if err = r.CompleteMaintenance(context.Background(), j.ID, true, "", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
}
func TestMaintenanceRetry(t *testing.T) {
	r := newRepo(t)
	now := time.Now().UTC()
	j, _ := r.QueueMaintenance(context.Background(), 2, now)
	_, _ = r.ClaimMaintenance(context.Background(), now.Add(time.Second))
	if e := r.CompleteMaintenance(context.Background(), j.ID, false, "sensor timeout", now); e != nil {
		t.Fatal(e)
	}
	var status string
	if e := r.DB().QueryRow("SELECT status FROM maintenance_jobs WHERE id=?", j.ID).Scan(&status); e != nil || status != "retry" {
		t.Fatalf("%s %v", status, e)
	}
}
func TestCheckinInsertIgnore(t *testing.T) {
	r := newRepo(t)
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		if e := r.RecordCheckin(context.Background(), 1, "opening", "key", now); e != nil {
			t.Fatal(e)
		}
	}
	var n int
	_ = r.DB().QueryRow("SELECT COUNT(*) FROM checkins WHERE idempotency_key='key'").Scan(&n)
	if n != 1 {
		t.Fatalf("%d", n)
	}
}
func TestAuditEvent(t *testing.T) {
	r := newRepo(t)
	e := r.AddAudit(context.Background(), domain.AuditEvent{ActorUserID: 1, EntityType: "stay", EntityID: "1", Action: "open", Result: "success", RequestID: "r", CreatedAt: time.Now().UTC()})
	if e != nil {
		t.Fatal(e)
	}
	var n int
	_ = r.DB().QueryRow("SELECT COUNT(*) FROM audit_events").Scan(&n)
	if n != 1 {
		t.Fatal(n)
	}
}
