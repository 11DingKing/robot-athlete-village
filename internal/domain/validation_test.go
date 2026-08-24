package domain

import (
	"testing"
	"time"
)

func TestAdmissionValidation(t *testing.T) {
	cases := []struct {
		name  string
		req   AdmissionRequest
		valid bool
	}{{"valid", AdmissionRequest{DelegationID: 1, RoomID: 2, IdempotencyKey: "key"}, true}, {"delegation", AdmissionRequest{DelegationID: 0, RoomID: 2, IdempotencyKey: "key"}, false}, {"room", AdmissionRequest{DelegationID: 1, RoomID: 0, IdempotencyKey: "key"}, false}, {"empty", AdmissionRequest{DelegationID: 1, RoomID: 2}, false}, {"spaces", AdmissionRequest{DelegationID: 1, RoomID: 2, IdempotencyKey: "   "}, false}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.req.Validate() == nil; got != tc.valid {
				t.Fatalf("got %v", got)
			}
		})
	}
}
func TestBookingValidation(t *testing.T) {
	cases := []struct {
		name  string
		req   BookingRequest
		valid bool
	}{{"valid", BookingRequest{AthleteID: 1, SlotID: 1, CoachID: 2, IdempotencyKey: "key"}, true}, {"athlete", BookingRequest{SlotID: 1, CoachID: 2, IdempotencyKey: "key"}, false}, {"slot", BookingRequest{AthleteID: 1, CoachID: 2, IdempotencyKey: "key"}, false}, {"coach", BookingRequest{AthleteID: 1, SlotID: 1, IdempotencyKey: "key"}, false}, {"key", BookingRequest{AthleteID: 1, SlotID: 1, CoachID: 2}, false}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.req.Validate() == nil; got != tc.valid {
				t.Fatalf("got %v", got)
			}
		})
	}
}
func TestTimeWindow(t *testing.T) {
	base := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	valid := TimeWindow{base, base.Add(time.Hour)}
	if !valid.Valid() {
		t.Fatal("valid rejected")
	}
	if (TimeWindow{}).Valid() {
		t.Fatal("zero valid")
	}
	if valid.Overlaps(TimeWindow{base.Add(30 * time.Minute), base.Add(90 * time.Minute)}) != true {
		t.Fatal("overlap")
	}
	if valid.Overlaps(TimeWindow{base.Add(time.Hour), base.Add(2 * time.Hour)}) {
		t.Fatal("touch overlap")
	}
	if valid.Overlaps(TimeWindow{base.Add(2 * time.Hour), base.Add(3 * time.Hour)}) {
		t.Fatal("disjoint overlap")
	}
}
func TestNormalizeName(t *testing.T) {
	cases := map[string]string{"  north   arena ": "north arena", "": "", "one\ttwo": "one two", "  多功能  场  ": "多功能 场"}
	for in, want := range cases {
		if got := NormalizeName(in); got != want {
			t.Fatalf("%q => %q", in, got)
		}
	}
}
func TestSupportedCategories(t *testing.T) {
	for _, v := range []string{"mobility", "balance", "rescue", "relay", " MOBILITY "} {
		if !IsSupportedCategory(v) {
			t.Fatalf("reject %s", v)
		}
	}
	for _, v := range []string{"", "unknown", "speed"} {
		if IsSupportedCategory(v) {
			t.Fatalf("accept %s", v)
		}
	}
}
func TestSupportedRoles(t *testing.T) {
	if !IsSupportedRole(RoleAdmin) || !IsSupportedRole(RoleCoach) || IsSupportedRole(Role("guest")) {
		t.Fatal("roles")
	}
}
func TestBatchSize(t *testing.T) {
	if ValidateBatchSize(0) == nil {
		t.Fatal("empty accepted")
	}
	if ValidateBatchSize(101) == nil {
		t.Fatal("large accepted")
	}
	if ValidateBatchSize(2) != nil {
		t.Fatal("valid rejected")
	}
}
