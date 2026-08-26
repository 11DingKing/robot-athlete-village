package domain

import "testing"

func TestStayTransitions(t *testing.T) {
	cases := []struct {
		name     string
		from, to StayStatus
		want     bool
	}{
		{"pending-active", StayPending, StayActive, true}, {"active-closed", StayActive, StayClosed, true}, {"pending-closed", StayPending, StayClosed, false}, {"closed-active", StayClosed, StayActive, false}, {"closed-pending", StayClosed, StayPending, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.from.CanTransition(tc.to); got != tc.want {
				t.Fatalf("got %v", got)
			}
		})
	}
}
func TestBookingTransitions(t *testing.T) {
	cases := []struct {
		name     string
		from, to BookingStatus
		want     bool
	}{
		{"hold-confirm", BookingHeld, BookingConfirmed, true}, {"hold-cancel", BookingHeld, BookingCancelled, true}, {"confirm-cancel", BookingConfirmed, BookingCancelled, true}, {"confirm-hold", BookingConfirmed, BookingHeld, false}, {"cancel-confirm", BookingCancelled, BookingConfirmed, false}, {"cancel-hold", BookingCancelled, BookingHeld, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.from.CanTransition(tc.to); got != tc.want {
				t.Fatalf("got %v", got)
			}
		})
	}
}
func TestEquipmentTransitions(t *testing.T) {
	cases := []struct {
		name     string
		from, to EquipmentStatus
		want     bool
	}{
		{"ready-use", EquipmentReady, EquipmentInUse, true}, {"use-ready", EquipmentInUse, EquipmentReady, true}, {"ready-maint", EquipmentReady, EquipmentMaintenance, true}, {"maint-ready", EquipmentMaintenance, EquipmentReady, true}, {"use-maint", EquipmentInUse, EquipmentMaintenance, false}, {"maint-use", EquipmentMaintenance, EquipmentInUse, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.from.CanTransition(tc.to); got != tc.want {
				t.Fatalf("got %v", got)
			}
		})
	}
}
func TestRoleValues(t *testing.T) {
	if RoleAdmin == RoleCoach {
		t.Fatal("roles must differ")
	}
	if string(RoleAdmin) != "village_admin" {
		t.Fatal("admin value")
	}
	if string(RoleCoach) != "coach" {
		t.Fatal("coach value")
	}
}
func TestStatusesAreStable(t *testing.T) {
	values := []string{string(StayPending), string(StayActive), string(StayClosed), string(BookingHeld), string(BookingConfirmed), string(BookingCancelled), string(EquipmentReady), string(EquipmentInUse), string(EquipmentMaintenance)}
	seen := map[string]bool{}
	for _, v := range values {
		if seen[v] {
			t.Fatalf("duplicate %s", v)
		}
		seen[v] = true
	}
}
