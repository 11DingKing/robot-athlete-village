package service

import (
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"testing"
	"time"
)

func TestDefaultPolicy(t *testing.T) {
	p := DefaultPolicy()
	if p.MaxActiveBookings != 3 || len(p.AllowedCategories) != 4 {
		t.Fatalf("%+v", p)
	}
}
func TestPolicyAthlete(t *testing.T) {
	p := DefaultPolicy()
	valid := domain.Athlete{DelegationID: 1, DisplayName: "小虎机器人", Category: "mobility"}
	if e := p.ValidateAthlete(valid); e != nil {
		t.Fatal(e)
	}
	for _, a := range []domain.Athlete{{DelegationID: 1, DisplayName: "", Category: "mobility"}, {DelegationID: 1, DisplayName: "x", Category: "unknown"}, {DelegationID: 0, DisplayName: "x", Category: "mobility"}} {
		if p.ValidateAthlete(a) == nil {
			t.Fatalf("accepted %+v", a)
		}
	}
}
func TestPolicyWindows(t *testing.T) {
	p := DefaultPolicy()
	base := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	if !p.CheckInAllowed(base, base.Add(time.Hour)) || p.CheckInAllowed(base, base.Add(3*time.Hour)) {
		t.Fatal("checkin")
	}
	if !p.CheckOutAllowed(base, base.Add(24*time.Hour)) || p.CheckOutAllowed(base, base.Add(time.Hour)) {
		t.Fatal("checkout")
	}
}
func TestPolicyClone(t *testing.T) {
	p := DefaultPolicy()
	q := p.Clone()
	q.AllowedCategories["new"] = true
	if p.AllowedCategories["new"] {
		t.Fatal("shared map")
	}
}
func TestBookingLimit(t *testing.T) {
	p := DefaultPolicy()
	for i := 0; i < 4; i++ {
		if p.CanAddBooking(i) != (i < 3) {
			t.Fatalf("%d", i)
		}
	}
}
