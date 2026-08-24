package repository

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/pagination"
	"testing"
)

func TestDelegationQueries(t *testing.T) {
	r := newRepo(t)
	all, e := r.ListDelegations(context.Background(), pagination.Normalize(10, 0), "")
	if e != nil || all.Total != 2 {
		t.Fatalf("%+v %v", all, e)
	}
	approved, e := r.ListDelegations(context.Background(), pagination.Normalize(10, 0), "approved")
	if e != nil || approved.Total != 2 {
		t.Fatalf("%+v %v", approved, e)
	}
	if r.CountAthletes(context.Background(), 1) != 2 || !r.DelegationReady(context.Background(), 1) {
		t.Fatal("delegation facts")
	}
}
func TestDelegationStatusUpdate(t *testing.T) {
	r := newRepo(t)
	if e := r.UpdateDelegationStatus(context.Background(), 1, "approved", "active"); e != nil {
		t.Fatal(e)
	}
	if e := r.UpdateDelegationStatus(context.Background(), 1, "approved", "active"); e == nil {
		t.Fatal("stale update")
	}
	if e := r.UpdateDelegationStatus(context.Background(), 1, "active", "active"); e == nil {
		t.Fatal("same update")
	}
}
