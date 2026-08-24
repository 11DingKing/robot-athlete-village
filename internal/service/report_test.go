package service

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/pagination"
	"testing"
)

func TestOccupancyPagination(t *testing.T) {
	_, store := villageService(t)
	r := NewReport(store.DB())
	out, e := r.Occupancy(context.Background(), pagination.Normalize(2, 0))
	if e != nil || len(out.Items) != 2 || out.Total < 3 {
		t.Fatalf("%+v %v", out, e)
	}
}
func TestOccupancyOffset(t *testing.T) {
	_, store := villageService(t)
	r := NewReport(store.DB())
	a, _ := r.Occupancy(context.Background(), pagination.Normalize(1, 0))
	b, _ := r.Occupancy(context.Background(), pagination.Normalize(1, 1))
	if len(a.Items) != 1 || len(b.Items) != 1 || a.Items[0].RoomCode == b.Items[0].RoomCode {
		t.Fatalf("a=%+v b=%+v", a, b)
	}
}
func TestOccupancyEmptyPage(t *testing.T) {
	_, store := villageService(t)
	r := NewReport(store.DB())
	out, e := r.Occupancy(context.Background(), pagination.Normalize(20, 99))
	if e != nil || len(out.Items) != 0 || out.Total == 0 {
		t.Fatalf("%+v %v", out, e)
	}
}
