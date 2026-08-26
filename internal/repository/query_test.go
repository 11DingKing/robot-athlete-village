package repository

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/pagination"
	"testing"
)

func TestListSlotsAll(t *testing.T) {
	r := newRepo(t)
	out, e := r.ListSlots(context.Background(), pagination.Normalize(10, 0), "")
	if e != nil || len(out.Items) != 3 || out.Total != 3 {
		t.Fatalf("%+v %v", out, e)
	}
}
func TestListSlotsFilter(t *testing.T) {
	r := newRepo(t)
	for _, sport := range []string{"mobility", "balance", "missing"} {
		out, e := r.ListSlots(context.Background(), pagination.Normalize(10, 0), sport)
		if e != nil {
			t.Fatal(e)
		}
		if sport == "mobility" && (out.Total != 2 || len(out.Items) != 2) {
			t.Fatalf("mobility %+v", out)
		}
		if sport == "balance" && (out.Total != 1 || len(out.Items) != 1) {
			t.Fatalf("balance %+v", out)
		}
		if sport == "missing" && out.Total != 0 {
			t.Fatalf("missing %+v", out)
		}
	}
}
func TestSlotStatusTransition(t *testing.T) {
	r := newRepo(t)
	if e := r.SetSlotStatus(context.Background(), 1, "open", "closed"); e != nil {
		t.Fatal(e)
	}
	if e := r.SetSlotStatus(context.Background(), 1, "open", "closed"); e == nil {
		t.Fatal("stale transition")
	}
}
func TestFindAthlete(t *testing.T) {
	r := newRepo(t)
	a, e := r.FindAthlete(context.Background(), 1)
	if e != nil || a.DisplayName == "" {
		t.Fatalf("%+v %v", a, e)
	}
	if _, e = r.FindAthlete(context.Background(), 999); e == nil {
		t.Fatal("missing athlete")
	}
}
